import { query } from "@/lib/db";
import { computeTyfcbScore } from "@/lib/scoring";

export type TyfcbEntryStatus = "pending" | "verified" | "rejected";

export type PairPenaltyImpactRow = {
  entry_id: string;
  member_id: string;
  team_id: string | null;
  member_name: string;
  nilai: number;
  pair_ordinal: number;
  status: TyfcbEntryStatus;
  skor_lama: number;
  skor_baru: number;
  selisih: number;
};

export type PairPenaltyImpact = {
  rows: PairPenaltyImpactRow[];
  total_lama: number;
  total_baru: number;
  total_selisih: number;
  /** Portion of total_selisih that moves the leaderboard now — verified entries only. */
  ledger_selisih: number;
  verified_count: number;
  unverified_count: number;
  affected_members: number;
};

export async function getActiveSeason(): Promise<{ id: string; nama: string; pair_penalty_enabled: boolean } | null> {
  const { rows } = await query<{ id: string; nama: string; pair_penalty_enabled: boolean }>(
    `select id, nama, pair_penalty_enabled from event_seasons where nama = 'BRAG 2026' limit 1`
  );
  return rows[0] ?? null;
}

export async function getPairPenaltyEnabled(seasonId: string): Promise<boolean> {
  const { rows } = await query<{ pair_penalty_enabled: boolean }>(
    `select pair_penalty_enabled from event_seasons where id = $1 limit 1`,
    [seasonId]
  );
  return rows[0]?.pair_penalty_enabled ?? true;
}

export type BoosterMultiplier = {
  multiplier: number;
  /** Null when no booster applies — the caller records M = 1.0. */
  booster_id: string | null;
  judul: string | null;
};

/**
 * The M factor for one TYFCB entry (spec §4.1 Step 3).
 *
 * A booster applies when it is aktif, the transaction date falls inside its range, and
 * the transaction value falls inside its band (band_min inclusive, band_max exclusive —
 * same convention as TYFCB_BANDS; a NULL bound means unbounded on that side).
 *
 * Overlapping boosters resolve to the single highest multiplier. They are never
 * multiplied together, so stacked boosters cannot inflate a score without an admin
 * explicitly creating one large multiplier.
 */
export async function getBoosterMultiplier(
  seasonId: string,
  tanggal: string,
  nilai: number
): Promise<BoosterMultiplier> {
  const { rows } = await query<{ id: string; judul: string; multiplier: string }>(`
    select id, judul, multiplier
    from booster_events
    where season_id = $1
      and status = 'aktif'
      and $2::date between tanggal_mulai and tanggal_berakhir
      and (band_min is null or $3::bigint >= band_min)
      and (band_max is null or $3::bigint <  band_max)
    order by multiplier desc, created_at desc
    limit 1
  `, [seasonId, tanggal, nilai]);

  const top = rows[0];
  const multiplier = Number(top?.multiplier ?? 1);

  if (!top || !Number.isFinite(multiplier) || multiplier <= 1) {
    return { multiplier: 1, booster_id: null, judul: null };
  }

  return { multiplier, booster_id: top.id, judul: top.judul };
}

type ScorableEntry = {
  id: string;
  member_id: string;
  team_id: string | null;
  member_name: string;
  nilai: string;
  pair_ordinal: number | null;
  status: TyfcbEntryStatus;
  computed_score: number | null;
  event_multiplier_applied: string | null;
};

/**
 * What every entry would be worth if the season ran with `targetEnabled`.
 * Only entries whose score actually changes are returned — that is the set an apply would correct.
 *
 * Pending and rejected entries are included on purpose: their computed_score is what
 * gets written to the ledger the moment an admin verifies them, so leaving it stale
 * would award penalised points under a season that has the penalty switched off.
 */
export async function computePairPenaltyImpact(
  seasonId: string,
  targetEnabled: boolean
): Promise<PairPenaltyImpact> {
  const { rows: entries } = await query<ScorableEntry>(`
    select te.id, m.id as member_id, m.team_id, u.full_name as member_name,
           te.nilai::text as nilai,
           te.pair_ordinal, te.status::text as status,
           te.computed_score, te.event_multiplier_applied::text
    from tyfcb_entries te
    join members m on m.id = te.giver_id
    join app_users u on u.id = m.user_id
    where te.season_id = $1
    order by te.created_at
  `, [seasonId]);

  const rows: PairPenaltyImpactRow[] = [];

  for (const entry of entries) {
    const skorLama = entry.computed_score ?? 0;
    const skorBaru = computeTyfcbScore(Number(entry.nilai), entry.pair_ordinal ?? 1, {
      pairPenaltyEnabled: targetEnabled,
      eventMultiplier: Number(entry.event_multiplier_applied ?? 1) || 1,
    });

    if (skorBaru === skorLama) continue;

    rows.push({
      entry_id: entry.id,
      member_id: entry.member_id,
      team_id: entry.team_id,
      member_name: entry.member_name,
      nilai: Number(entry.nilai),
      pair_ordinal: entry.pair_ordinal ?? 1,
      status: entry.status,
      skor_lama: skorLama,
      skor_baru: skorBaru,
      selisih: skorBaru - skorLama,
    });
  }

  const verifiedFirst = (r: PairPenaltyImpactRow) => (r.status === "verified" ? 0 : 1);
  rows.sort(
    (a, b) =>
      verifiedFirst(a) - verifiedFirst(b) ||
      b.selisih - a.selisih ||
      a.member_name.localeCompare(b.member_name)
  );

  const verified = rows.filter((r) => r.status === "verified");

  return {
    rows,
    total_lama: rows.reduce((sum, r) => sum + r.skor_lama, 0),
    total_baru: rows.reduce((sum, r) => sum + r.skor_baru, 0),
    total_selisih: rows.reduce((sum, r) => sum + r.selisih, 0),
    ledger_selisih: verified.reduce((sum, r) => sum + r.selisih, 0),
    verified_count: verified.length,
    unverified_count: rows.length - verified.length,
    affected_members: new Set(rows.map((r) => r.member_id)).size,
  };
}
