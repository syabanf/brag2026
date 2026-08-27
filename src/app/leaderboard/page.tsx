import { AppShell } from "@/components/app-shell";
import { requireUser } from "@/lib/auth";
import { query } from "@/lib/db";
import { LeaderboardClient, type MemberRow, type TeamRow } from "./leaderboard-client";

async function getSeasonId(): Promise<string | null> {
  const { rows } = await query<{ id: string }>(
    `select id from event_seasons where nama = 'BRAG 2026' limit 1`,
    []
  );
  return rows[0]?.id ?? null;
}

async function getTeamStandings(seasonId: string): Promise<TeamRow[]> {
  const { rows } = await query<TeamRow>(`
    with tyfcb_by_team as (
      select m.team_id,
             coalesce(sum(te.nilai), 0)::bigint as nilai_tyfcb,
             count(te.id)::int                  as count_tyfcb
      from tyfcb_entries te
      join members m on m.id = te.giver_id and m.season_id = $1
      where te.status = 'verified'
      group by m.team_id
    ),
    visitor_by_team as (
      select m.team_id, count(v.id)::int as count_visitor
      from visitors v
      join members m on m.id = v.inviter_id and m.season_id = $1
      where v.is_void = false
      group by m.team_id
    )
    select
      t.id                           as team_id,
      t.nama_tim,
      coalesce(sum(sl.points), 0)::int as score_overall,
      coalesce(tt.nilai_tyfcb, 0)    as nilai_tyfcb,
      coalesce(tt.count_tyfcb, 0)    as count_tyfcb,
      coalesce(vt.count_visitor, 0)  as count_visitor
    from teams t
    left join score_ledger sl on sl.team_id = t.id and sl.season_id = $1
    left join tyfcb_by_team tt on tt.team_id = t.id
    left join visitor_by_team vt on vt.team_id = t.id
    where t.season_id = $1
    group by t.id, t.nama_tim, tt.nilai_tyfcb, tt.count_tyfcb, vt.count_visitor
    order by score_overall desc, substring(t.nama_tim, 5)::int
  `, [seasonId]);
  return rows;
}

async function getMemberStandings(seasonId: string): Promise<MemberRow[]> {
  const { rows } = await query<MemberRow>(`
    select
      m.id,
      u.full_name,
      t.nama_tim,
      c.nama                                                                        as klasifikasi_nama,
      m.color_status,
      coalesce(sum(sl.points), 0)::int                                             as score_overall,
      coalesce(sum(case when sl.kategori = 'tyfcb'   then sl.points end), 0)::int as score_tyfcb,
      coalesce(sum(case when sl.kategori = 'visitor'  then sl.points end), 0)::int as score_visitor
    from members m
    join app_users u on u.id = m.user_id
    left join teams t on t.id = m.team_id
    left join classifications c on c.id = m.klasifikasi_id
    left join score_ledger sl on sl.member_id = m.id and sl.season_id = $1
    where m.season_id = $1
    group by m.id, u.full_name, t.nama_tim, c.nama, m.color_status
    order by score_overall desc, u.full_name
  `, [seasonId]);
  return rows;
}

export default async function LeaderboardPage() {
  await requireUser();

  const seasonId = await getSeasonId();

  const [teams, members] = seasonId
    ? await Promise.all([getTeamStandings(seasonId), getMemberStandings(seasonId)])
    : [[], []];

  return (
    <AppShell>
      <div className="mb-6 flex flex-col justify-between gap-3 sm:flex-row sm:items-end">
        <div>
          <h1 className="text-3xl font-black text-ink">Leaderboard</h1>
          <p className="mt-2 text-muted">
            Pantau posisi timmu dan kontribusimu di tiap kategori — poin keseluruhan,
            nilai TYFCB kolektif, dan jumlah visitor yang berhasil diundang.
          </p>
        </div>
      </div>

      <LeaderboardClient teams={teams} members={members} />
    </AppShell>
  );
}
