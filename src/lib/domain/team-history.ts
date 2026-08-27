import { query } from "@/lib/db";
import type {
  TeamHistoryResponse,
  TeamTyfcbHistoryRow,
  TeamVisitorHistoryRow,
} from "@/lib/domain/types";

export const UUID_PATTERN =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

// Shared by the member and public history endpoints so the two can never drift.
// Only non-sensitive columns are selected — visitors.kontak is deliberately
// never returned, since this data also feeds the unauthenticated leaderboard.
export async function getTeamHistory(
  teamId: string
): Promise<TeamHistoryResponse | null> {
  const { rows: teamRows } = await query<{ nama_tim: string; season_id: string }>(`
    select t.nama_tim, t.season_id
    from teams t
    join event_seasons es on es.id = t.season_id and es.nama = 'BRAG 2026'
    where t.id = $1
    limit 1
  `, [teamId]);

  const team = teamRows[0];
  if (!team) return null;

  // Both queries mirror the leaderboard aggregation exactly, so the dialog list
  // always matches the count shown on the chip.
  const [tyfcbResult, visitorResult] = await Promise.all([
    query<TeamTyfcbHistoryRow>(`
      select
        te.id,
        u_buyer.full_name  as buyer_name,
        u_seller.full_name as seller_name,
        to_char(te.tanggal, 'DD Mon YYYY') as tanggal,
        te.computed_score,
        te.event_multiplier_applied::text as event_multiplier_applied,
        boost.judul as booster_judul
      from tyfcb_entries te
      join members m_buyer      on m_buyer.id     = te.giver_id
                               and m_buyer.season_id = $2
      join app_users u_buyer    on u_buyer.id     = m_buyer.user_id
      join members m_seller     on m_seller.id    = te.receiver_id
      join app_users u_seller   on u_seller.id    = m_seller.user_id
      -- Entries record the multiplier but not which booster produced it, so the title
      -- is re-derived with getBoosterMultiplier's rules. Null once an admin edits the
      -- booster out of range, which the badge falls back on gracefully.
      left join lateral (
        select be.judul
        from booster_events be
        where be.season_id = te.season_id
          and be.status = 'aktif'
          and be.multiplier > 1
          and te.tanggal between be.tanggal_mulai and be.tanggal_berakhir
          and (be.band_min is null or te.nilai >= be.band_min)
          and (be.band_max is null or te.nilai <  be.band_max)
        order by be.multiplier desc, be.created_at desc
        limit 1
      ) boost on true
      where m_buyer.team_id = $1
        and te.status = 'verified'
      order by te.tanggal desc, te.created_at desc
    `, [teamId, team.season_id]),
    query<TeamVisitorHistoryRow>(`
      select
        v.id,
        v.nama,
        u.full_name as inviter_name,
        to_char(v.tanggal_undang, 'DD Mon YYYY') as tanggal_undang,
        v.status_hadir::text as status_hadir,
        v.is_converted
      from visitors v
      join members m   on m.id = v.inviter_id and m.season_id = $2
      join app_users u on u.id = m.user_id
      where m.team_id = $1
        and v.is_void = false
      order by v.tanggal_undang desc, v.created_at desc
    `, [teamId, team.season_id]),
  ]);

  return {
    nama_tim: team.nama_tim,
    tyfcb: tyfcbResult.rows,
    visitors: visitorResult.rows,
  };
}
