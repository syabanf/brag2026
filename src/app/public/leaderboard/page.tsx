import { query } from "@/lib/db";
import Image from "next/image";
import { CopyLinkButton } from "./copy-link-button";
import { TeamHistoryChips, TeamHistoryInline } from "./team-history-chips";

type TeamRow = {
  team_id: string;
  nama_tim: string;
  score_overall: number;
  nilai_tyfcb: number;
  count_tyfcb: number;
  count_visitor: number;
};

async function getTeamStandings(): Promise<TeamRow[]> {
  const { rows } = await query<TeamRow>(`
    with tyfcb_by_team as (
      select m.team_id,
             coalesce(sum(te.nilai), 0)::bigint as nilai_tyfcb,
             count(te.id)::int                  as count_tyfcb
      from tyfcb_entries te
      join members m on m.id = te.giver_id
      join event_seasons es on es.id = m.season_id and es.nama = 'BRAG 2026'
      where te.status = 'verified'
      group by m.team_id
    ),
    visitor_by_team as (
      select m.team_id, count(v.id)::int as count_visitor
      from visitors v
      join members m on m.id = v.inviter_id
      join event_seasons es on es.id = m.season_id and es.nama = 'BRAG 2026'
      where v.is_void = false
      group by m.team_id
    )
    select
      t.id                             as team_id,
      t.nama_tim,
      coalesce(sum(sl.points), 0)::int as score_overall,
      coalesce(tt.nilai_tyfcb, 0)      as nilai_tyfcb,
      coalesce(tt.count_tyfcb, 0)      as count_tyfcb,
      coalesce(vt.count_visitor, 0)    as count_visitor
    from teams t
    join event_seasons es on es.id = t.season_id and es.nama = 'BRAG 2026'
    left join score_ledger sl     on sl.team_id = t.id and sl.season_id = es.id
    left join tyfcb_by_team tt    on tt.team_id = t.id
    left join visitor_by_team vt  on vt.team_id = t.id
    group by t.id, t.nama_tim, tt.nilai_tyfcb, tt.count_tyfcb, vt.count_visitor
    order by score_overall desc, substring(t.nama_tim, 5)::int
  `, []);
  return rows;
}

const RANK_CARD = [
  {
    card:    "bg-gradient-to-br from-yellow-300 via-amber-400 to-orange-400 shadow-xl",
    badge:   "bg-white/40 text-yellow-900",
    name:    "text-yellow-900",
    score:   "text-yellow-900",
    chip:    "bg-white/30 text-yellow-900",
    label:   "🥇 Juara 1",
    labelCls:"text-yellow-800/70",
  },
  {
    card:    "bg-gradient-to-br from-slate-300 via-gray-200 to-slate-400 shadow-lg",
    badge:   "bg-white/40 text-slate-700",
    name:    "text-slate-800",
    score:   "text-slate-800",
    chip:    "bg-white/30 text-slate-700",
    label:   "🥈 Juara 2",
    labelCls:"text-slate-600/70",
  },
  {
    card:    "bg-gradient-to-br from-amber-700 via-amber-600 to-orange-700 shadow-lg",
    badge:   "bg-white/25 text-white",
    name:    "text-white",
    score:   "text-white",
    chip:    "bg-white/20 text-white/90",
    label:   "🥉 Juara 3",
    labelCls:"text-white/70",
  },
];

export const dynamic = "force-dynamic";

export default async function PublicLeaderboardPage() {
  const teams = await getTeamStandings();
  const top3  = teams.slice(0, 3);
  const rest  = teams.slice(3);
  const updatedAt = new Date().toLocaleString("id-ID", {
    dateStyle: "medium", timeStyle: "short", timeZone: "Asia/Jakarta",
  });

  return (
    <div className="min-h-screen bg-gradient-to-b from-red-50/60 via-white to-rose-50/30">
      {/* Header */}
      <div className="sticky top-0 z-10 border-b border-red-100 bg-white/80 backdrop-blur-md">
        <div className="mx-auto flex max-w-lg items-center justify-between px-4 py-3">
          <Image
            src="/bni-grow-logo.png"
            alt="BNI Grow"
            width={120}
            height={37}
            className="h-9 w-auto object-contain"
            priority
          />
          <CopyLinkButton />
        </div>
      </div>

      <div className="mx-auto max-w-lg px-4 py-8">
        {/* Title */}
        <div className="mb-8 text-center">
          <h1 className="text-3xl font-black text-gray-900">Team Leaderboard</h1>
          <p className="mt-1 text-sm font-semibold text-gray-500">Overall Season BRAG 2026</p>
          <div className="mt-3 flex items-center justify-center gap-1.5">
            <span className="relative flex h-2 w-2">
              <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-400 opacity-75" />
              <span className="relative inline-flex h-2 w-2 rounded-full bg-green-500" />
            </span>
            <span className="text-xs font-semibold text-gray-400">Live · {updatedAt} WIB</span>
          </div>
        </div>

        {/* Top 3 */}
        <div className="mb-4 space-y-3">
          {top3.map((team, i) => {
            const s = RANK_CARD[i];
            return (
              <div key={team.team_id} className={`rounded-2xl p-5 ${s.card}`}>
                <div className="flex items-center justify-between gap-3">
                  <div className="flex items-center gap-4">
                    <span className={`grid h-14 w-14 shrink-0 place-items-center rounded-full text-2xl font-black ${s.badge}`}>
                      {i + 1}
                    </span>
                    <div>
                      <p className={`text-[0.65rem] font-bold uppercase tracking-widest ${s.labelCls}`}>{s.label}</p>
                      <p className={`text-xl font-black ${s.name}`}>{team.nama_tim}</p>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className={`text-2xl font-black ${s.score}`}>
                      {Number(team.score_overall).toLocaleString("id-ID")}
                    </p>
                    <p className={`text-xs font-bold ${s.labelCls}`}>pts</p>
                  </div>
                </div>
                <TeamHistoryChips
                  teamId={team.team_id}
                  teamName={team.nama_tim}
                  nilaiTyfcb={Number(team.nilai_tyfcb)}
                  countTyfcb={team.count_tyfcb}
                  countVisitor={team.count_visitor}
                  chipClassName={s.chip}
                />
              </div>
            );
          })}
        </div>

        {/* Rank 4+ */}
        {rest.length > 0 && (
          <div className="overflow-hidden rounded-2xl border border-red-100 bg-white shadow-sm">
            {rest.map((team, i) => (
              <div
                key={team.team_id}
                className="flex items-center justify-between gap-3 border-b border-red-50 px-4 py-4 last:border-0"
              >
                <div className="flex items-center gap-4">
                  <span className="grid h-9 w-9 shrink-0 place-items-center rounded-full bg-gray-100 text-sm font-black text-gray-500">
                    {i + 4}
                  </span>
                  <div>
                    <p className="font-black text-gray-900">{team.nama_tim}</p>
                    <TeamHistoryInline
                      teamId={team.team_id}
                      teamName={team.nama_tim}
                      nilaiTyfcb={Number(team.nilai_tyfcb)}
                      countTyfcb={team.count_tyfcb}
                      countVisitor={team.count_visitor}
                    />
                  </div>
                </div>
                <p className="text-lg font-black text-gray-700">
                  {Number(team.score_overall).toLocaleString("id-ID")} <span className="text-xs font-semibold text-gray-400">pts</span>
                </p>
              </div>
            ))}
          </div>
        )}

        <p className="mt-10 text-center text-xs text-gray-300">
          BRAG 2026 · BNI Grow Annual Challenge
        </p>
      </div>
    </div>
  );
}
