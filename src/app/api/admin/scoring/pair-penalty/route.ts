import { NextRequest, NextResponse } from "next/server";
import { requireUser } from "@/lib/auth";
import { withTransaction } from "@/lib/db";
import { computePairPenaltyImpact, getActiveSeason } from "@/lib/scoring-settings";

async function requireAdminUser() {
  const { user } = await requireUser();
  if (user.role !== "admin" && user.role !== "super_admin") return null;
  return user;
}

/** Current setting plus what would change if it were flipped. */
export async function GET() {
  const user = await requireAdminUser();
  if (!user) return NextResponse.json({ error: "Forbidden" }, { status: 403 });

  const season = await getActiveSeason();
  if (!season) return NextResponse.json({ error: "Season aktif tidak ditemukan." }, { status: 404 });

  const impact = await computePairPenaltyImpact(season.id, !season.pair_penalty_enabled);

  return NextResponse.json({
    enabled: season.pair_penalty_enabled,
    season_nama: season.nama,
    impact,
  });
}

export async function PATCH(req: NextRequest) {
  const user = await requireAdminUser();
  if (!user) return NextResponse.json({ error: "Forbidden" }, { status: 403 });

  const body = await req.json();
  const { enabled, apply_corrections: applyCorrections = false } = body as {
    enabled?: unknown;
    apply_corrections?: unknown;
  };

  if (typeof enabled !== "boolean") {
    return NextResponse.json({ error: "Field 'enabled' harus boolean." }, { status: 400 });
  }
  if (typeof applyCorrections !== "boolean") {
    return NextResponse.json({ error: "Field 'apply_corrections' harus boolean." }, { status: 400 });
  }

  const season = await getActiveSeason();
  if (!season) return NextResponse.json({ error: "Season aktif tidak ditemukan." }, { status: 404 });

  // Impact is computed against the CURRENT stored scores, before the setting flips.
  const impact = await computePairPenaltyImpact(season.id, enabled);

  const corrected = await withTransaction(async (q) => {
    await q(`update event_seasons set pair_penalty_enabled = $1 where id = $2`, [
      enabled,
      season.id,
    ]);

    let entries = 0;
    let ledger = 0;

    for (const row of impact.rows) {
      const isVerified = row.status === "verified";

      // A pending or rejected entry has never scored — its computed_score is only the
      // preview that verification will later push to the ledger, so it always follows
      // the season setting. Verified entries already scored, so restating them is the
      // admin's call via apply_corrections.
      if (isVerified && !applyCorrections) continue;

      await q(`update tyfcb_entries set computed_score = $1 where id = $2`, [
        row.skor_baru,
        row.entry_id,
      ]);
      entries += 1;

      if (!isVerified) continue;

      // score_ledger is append-only: record the delta, never rewrite the original row.
      await q(
        `insert into score_ledger (season_id, member_id, team_id, kategori, points, sumber_ref, keterangan)
         values ($1, $2, $3, 'tyfcb', $4, $5, $6)`,
        [
          season.id,
          row.member_id,
          row.team_id,
          row.selisih,
          row.entry_id,
          `Koreksi pair penalty (${enabled ? "diaktifkan" : "dinonaktifkan"})`,
        ]
      );
      ledger += 1;
    }

    return { entries, ledger };
  });

  return NextResponse.json({
    ok: true,
    enabled,
    corrected_entries: corrected.entries,
    ledger_rows: corrected.ledger,
  });
}
