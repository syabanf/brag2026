import { NextRequest, NextResponse } from "next/server";
import { requireUser } from "@/lib/auth";
import { query } from "@/lib/db";
import { computeTyfcbScore } from "@/lib/scoring";
import { getPairPenaltyEnabled } from "@/lib/scoring-settings";

export async function POST(req: NextRequest) {
  const { user } = await requireUser();

  const body = await req.json();
  // buyer_id = the member who BOUGHT / gave business → gets the TYFCB points
  // current user (me) = seller who records the transaction (receiver_id)
  const { buyer_id, nilai: nilaiRaw, tanggal } = body;

  if (!buyer_id || !nilaiRaw || !tanggal) {
    return NextResponse.json({ error: "buyer_id, nilai, dan tanggal wajib diisi." }, { status: 400 });
  }

  const nilai = Number(nilaiRaw);
  if (isNaN(nilai) || nilai <= 0) {
    return NextResponse.json({ error: "Nilai tidak valid." }, { status: 400 });
  }

  // Seller = the logged-in user
  const { rows: memberRows } = await query<{ id: string; season_id: string }>(`
    select m.id, m.season_id
    from members m
    join event_seasons es on es.id = m.season_id
    where m.user_id = $1 and es.nama = 'BRAG 2026'
    limit 1
  `, [user.id]);

  if (!memberRows[0]) {
    return NextResponse.json({ error: "Profil member tidak ditemukan di season ini." }, { status: 404 });
  }

  const seller = memberRows[0];

  if (buyer_id === seller.id) {
    return NextResponse.json({ error: "Tidak bisa mencatat TYFCB ke diri sendiri." }, { status: 400 });
  }

  // Validate buyer exists in this season
  const { rows: buyerRows } = await query(
    `select id from members where id = $1 and season_id = $2 limit 1`,
    [buyer_id, seller.season_id]
  );
  if (!buyerRows[0]) {
    return NextResponse.json({ error: "Pembeli tidak ditemukan di season ini." }, { status: 400 });
  }

  // pair_ordinal: how many times this buyer has bought from this seller
  // giver_id = buyer (gets points), receiver_id = seller (records the transaction)
  const { rows: pairRows } = await query<{ count: string }>(
    `select count(*) as count from tyfcb_entries
     where giver_id = $1 and receiver_id = $2 and season_id = $3`,
    [buyer_id, seller.id, seller.season_id]
  );
  const pairOrdinal = Number(pairRows[0]?.count ?? 0) + 1;

  const M = 1.0;
  const pairPenaltyEnabled = await getPairPenaltyEnabled(seller.season_id);
  const computedScore = computeTyfcbScore(nilai, pairOrdinal, {
    pairPenaltyEnabled,
    eventMultiplier: M,
  });

  // giver_id = buyer (receives the TYFCB points)
  // receiver_id = seller (who inputted the form)
  const { rows: inserted } = await query<{ id: string }>(`
    insert into tyfcb_entries
      (season_id, giver_id, receiver_id, nilai, tanggal, status, computed_score, pair_ordinal, event_multiplier_applied)
    values ($1, $2, $3, $4, $5, 'pending', $6, $7, $8)
    returning id
  `, [seller.season_id, buyer_id, seller.id, nilai, tanggal, computedScore, pairOrdinal, M]);

  return NextResponse.json({ id: inserted[0].id, computed_score: computedScore, status: "pending" }, { status: 201 });
}
