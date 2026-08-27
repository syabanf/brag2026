import Link from "next/link";
import { Banknote, ChevronRight, Crown, Gift, Medal, Receipt, Trophy, UserPlus, Users, Zap } from "lucide-react";
import { AppShell } from "@/components/app-shell";
import { StatCard } from "@/components/stat-card";
import { requireUser } from "@/lib/auth";
import { query } from "@/lib/db";
import { formatCurrencyCompact, formatPoints } from "@/lib/utils";

// ─── DB query helpers ───────────────────────────────────────────────────────

async function getMemberProfile(userId: string) {
  const { rows } = await query<{
    id: string;
    color_status: string;
    team_id: string | null;
    nama_tim: string | null;
    klasifikasi_nama: string | null;
    full_name: string;
    season_id: string;
  }>(`
    select m.id, m.color_status, m.team_id, m.season_id,
           u.full_name,
           t.nama_tim,
           c.nama as klasifikasi_nama
    from members m
    join app_users u on u.id = m.user_id
    join event_seasons es on es.id = m.season_id
    left join teams t on t.id = m.team_id
    left join classifications c on c.id = m.klasifikasi_id
    where u.id = $1 and es.nama = 'BRAG 2026'
    limit 1
  `, [userId]);
  return rows[0] ?? null;
}

async function getTeamStandings(seasonId: string) {
  const { rows } = await query<{
    team_id: string;
    nama_tim: string;
    score_overall: number;
    nilai_tyfcb: number;
    count_tyfcb: number;
    count_visitor: number;
  }>(`
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
    left join score_ledger sl     on sl.team_id = t.id and sl.season_id = $1
    left join tyfcb_by_team tt    on tt.team_id = t.id
    left join visitor_by_team vt  on vt.team_id = t.id
    where t.season_id = $1
    group by t.id, t.nama_tim, tt.nilai_tyfcb, tt.count_tyfcb, vt.count_visitor
    order by score_overall desc, substring(t.nama_tim, 5)::int
  `, [seasonId]);
  return rows;
}

async function getActiveBoosters(seasonId: string) {
  const { rows } = await query<{
    id: string;
    judul: string;
    deskripsi: string | null;
    poin: number;
    tanggal_mulai: string;
    tanggal_berakhir: string;
  }>(`
    select id, judul, deskripsi, poin,
           to_char(tanggal_mulai,    'DD Mon YYYY') as tanggal_mulai,
           to_char(tanggal_berakhir, 'DD Mon YYYY') as tanggal_berakhir
    from booster_events
    where season_id = $1
      and status = 'aktif'
      and current_date between tanggal_mulai and tanggal_berakhir
    order by tanggal_mulai
  `, [seasonId]);
  return rows;
}

async function getTopContributors(seasonId: string) {
  const { rows } = await query<{
    kategori: string;
    full_name: string;
    nama_tim: string | null;
    score: number;
  }>(`
    with member_scores as (
      select
        m.id,
        u.full_name,
        t.nama_tim,
        coalesce(sum(case when sl.kategori = 'tyfcb'  then sl.points else 0 end), 0) as score_tyfcb,
        coalesce(sum(case when sl.kategori = 'visitor' then sl.points else 0 end), 0) as score_visitor,
        coalesce(sum(sl.points), 0) as score_overall
      from members m
      join app_users u on u.id = m.user_id
      left join teams t on t.id = m.team_id
      left join score_ledger sl on sl.member_id = m.id and sl.season_id = $1
      where m.season_id = $1
      group by m.id, u.full_name, t.nama_tim
    )
    select * from (
      select 'overall' as kategori, full_name, nama_tim, score_overall::int as score
      from member_scores order by score_overall desc limit 1
    ) t1
    union all
    select * from (
      select 'tyfcb', full_name, nama_tim, score_tyfcb::int
      from member_scores where score_tyfcb > 0 order by score_tyfcb desc limit 1
    ) t2
    union all
    select * from (
      select 'visitor', full_name, nama_tim, score_visitor::int
      from member_scores where score_visitor > 0 order by score_visitor desc limit 1
    ) t3
  `, [seasonId]);
  return rows;
}

async function getRecentTyfcb(memberId: string) {
  const { rows } = await query<{
    id: string;
    nilai: number;
    tanggal: string;
    status: string;
    computed_score: number | null;
    buyer_name: string | null;
  }>(`
    select
      te.id, te.nilai, to_char(te.tanggal, 'DD Mon YYYY') as tanggal, te.status, te.computed_score,
      u.full_name as buyer_name
    from tyfcb_entries te
    left join members mg on mg.id = te.giver_id
    left join app_users u on u.id = mg.user_id
    where te.receiver_id = $1
    order by te.created_at desc
    limit 5
  `, [memberId]);
  return rows;
}

// ─── Page ───────────────────────────────────────────────────────────────────

const quickActions = [
  { href: "/submit?type=tyfcb",   label: "TYFCB",  helper: "Closed business", icon: Gift },
  { href: "/submit?type=visitor", label: "Visitor", helper: "Undang tamu",     icon: UserPlus },
];

export default async function MemberDashboardPage() {
  const { user } = await requireUser();

  const member = await getMemberProfile(user.id);

  // If no member profile yet, show minimal dashboard
  if (!member) {
    return (
      <AppShell>
        <div className="flex flex-col items-center justify-center py-24 text-center">
          <p className="text-2xl font-black text-ink">Halo, {user.full_name.split(" ")[0]}.</p>
          <p className="mt-2 text-muted">Profil member kamu belum terdaftar di season BRAG 2026.</p>
          <p className="mt-1 text-sm text-muted">Hubungi Grow Coordinator untuk mendaftarkan kamu.</p>
        </div>
      </AppShell>
    );
  }

  // Run all remaining queries in parallel
  const [teams, activeBoosters, recentTyfcb, topContributors] = await Promise.all([
    getTeamStandings(member.season_id),
    getActiveBoosters(member.season_id),
    getRecentTyfcb(member.id),
    getTopContributors(member.season_id),
  ]);


  const myTeam = teams.find((t) => t.team_id === member.team_id);

  // Season-wide totals, derived from the standings already fetched above.
  const totalTransaksi = teams.reduce((sum, t) => sum + Number(t.count_tyfcb), 0);
  const totalNominal   = teams.reduce((sum, t) => sum + Number(t.nilai_tyfcb), 0);
  const totalVisitor   = teams.reduce((sum, t) => sum + Number(t.count_visitor), 0);

  return (
    <AppShell>
      <div className="grid gap-5 lg:grid-cols-[1fr_340px]">
        <section className="space-y-5">
          <div>
            <p className="section-label text-brand-700">
              Member Dashboard
            </p>
            <h1 className="mt-1.5 text-[1.75rem] font-black leading-tight tracking-tight text-ink sm:text-4xl">
              Halo, {member.full_name.split(" ")[0]}.
            </h1>
            <p className="mt-2 max-w-2xl text-sm leading-relaxed text-muted sm:text-base">
              Bantu sesama member closing dan undang tamu.{" "}
              {myTeam ? `Setiap kontribusi bergerak ke skor ${myTeam.nama_tim}.` : ""}
            </p>
          </div>

          {/* Row 1 = season-wide totals, row 2 = this member's team, aligned by category. */}
          <div className="grid grid-cols-2 gap-2.5 sm:gap-3 lg:grid-cols-3">
            <StatCard
              icon={Receipt}
              label="Total Transaksi"
              value={`${totalTransaksi.toLocaleString("id-ID")}×`}
              helper={`TYFCB verified dari ${teams.length} team`}
              tone="emerald"
            />
            <StatCard
              icon={Banknote}
              label="Total Nominal TYFCB"
              value={formatCurrencyCompact(totalNominal)}
              helper={`Akumulasi semua ${teams.length} team`}
              tone="sky"
            />
            <StatCard
              icon={Users}
              label="Total Visitor"
              value={`${totalVisitor.toLocaleString("id-ID")} tamu`}
              helper={`Akumulasi semua ${teams.length} team`}
              tone="amber"
            />
            <StatCard
              icon={Trophy}
              label={`Point ${myTeam?.nama_tim ?? "—"}`}
              value={`${formatPoints(Number(myTeam?.score_overall ?? 0))} pts`}
              helper="Total poin team keseluruhan"
              tone="brand"
            />
            <StatCard
              icon={Banknote}
              label="TYFCB Team"
              value={formatCurrencyCompact(Number(myTeam?.nilai_tyfcb ?? 0))}
              helper={`${Number(myTeam?.count_tyfcb ?? 0)}× transaksi verified`}
              tone="violet"
            />
            <StatCard
              icon={Users}
              label="Visitor Team"
              value={`${Number(myTeam?.count_visitor ?? 0)} tamu`}
              helper={`Akumulasi undangan ${myTeam?.nama_tim ?? "team"}`}
              tone="orange"
            />
          </div>

          {/* Quick actions */}
          <div className="grid gap-3 sm:grid-cols-2">
            {quickActions.map((action) => {
              const Icon = action.icon;
              return (
                <Link
                  href={action.href}
                  key={action.label}
                  className="group flex items-center gap-4 rounded-2xl border border-brand-200 bg-white px-4 py-4 shadow-sm transition hover:border-brand-400 hover:shadow-md active:scale-[0.98]"
                >
                  <span className="grid h-12 w-12 shrink-0 place-items-center rounded-full bg-brand-600 text-white shadow-sm transition group-hover:bg-brand-700">
                    <Icon className="h-6 w-6" />
                  </span>
                  <span className="flex-1">
                    <span className="block text-lg font-black text-ink">{action.label}</span>
                    <span className="text-sm font-medium text-muted">{action.helper}</span>
                  </span>
                  <ChevronRight className="h-5 w-5 shrink-0 text-brand-400 transition group-hover:translate-x-0.5 group-hover:text-brand-600" />
                </Link>
              );
            })}
          </div>

          {/* Top Contributors */}
          {topContributors.length > 0 && (() => {
            const TOP_META: Record<string, {
              label: string;
              sublabel: string;
              icon: React.ElementType;
              gradient: string;
              avatarBg: string;
              badgeBg: string;
            }> = {
              overall: {
                label: "Overall",
                sublabel: "Total Poin Tertinggi",
                icon: Crown,
                gradient: "from-yellow-400 via-amber-400 to-orange-400",
                avatarBg: "bg-white/25",
                badgeBg: "bg-white/20 text-white",
              },
              tyfcb: {
                label: "TYFCB",
                sublabel: "Poin TYFCB Tertinggi",
                icon: Medal,
                gradient: "from-brand-600 via-red-600 to-rose-500",
                avatarBg: "bg-white/20",
                badgeBg: "bg-white/20 text-white",
              },
              visitor: {
                label: "Visitor",
                sublabel: "Poin Visitor Tertinggi",
                icon: Trophy,
                gradient: "from-orange-500 via-amber-500 to-yellow-400",
                avatarBg: "bg-white/25",
                badgeBg: "bg-white/20 text-white",
              },
            };

            return (
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <p className="text-sm font-black uppercase tracking-[0.1em] text-brand-700">
                    ⭐ Top Kontributor
                  </p>
                  <Link href="/leaderboard" className="text-sm font-bold text-brand-600 hover:underline">
                    Lihat Leaderboard →
                  </Link>
                </div>
                <div className="grid grid-cols-3 gap-2.5 sm:gap-3">
                  {topContributors.map((top) => {
                    const meta = TOP_META[top.kategori];
                    if (!meta) return null;
                    const Icon = meta.icon;
                    const initials = top.full_name
                      .split(" ")
                      .map((n: string) => n[0])
                      .join("")
                      .slice(0, 2)
                      .toUpperCase();

                    return (
                      <div
                        key={top.kategori}
                        className={`relative overflow-hidden rounded-2xl bg-gradient-to-br ${meta.gradient} p-4 text-white shadow-lift`}
                      >
                        {/* Background decoration */}
                        <div className="pointer-events-none absolute -right-4 -top-4 h-24 w-24 rounded-full bg-white/10" />
                        <div className="pointer-events-none absolute -bottom-6 -left-6 h-20 w-20 rounded-full bg-white/10" />

                        {/* Category header */}
                        <div className="relative flex items-center gap-2 mb-3">
                          <span className="grid h-7 w-7 place-items-center rounded-full bg-white/20">
                            <Icon className="h-3.5 w-3.5 text-white" />
                          </span>
                          <div>
                            <p className="text-[0.65rem] font-bold uppercase tracking-[0.15em] text-white/70">
                              {meta.sublabel}
                            </p>
                            <p className="text-xs font-black text-white leading-tight">{meta.label}</p>
                          </div>
                        </div>

                        {/* Winner info */}
                        <div className="relative flex items-center gap-3">
                          <span className={`grid h-11 w-11 shrink-0 place-items-center rounded-full ${meta.avatarBg} text-base font-black text-white ring-2 ring-white/30`}>
                            {initials}
                          </span>
                          <div className="min-w-0 flex-1">
                            <p className="truncate font-black text-white leading-tight">
                              {top.full_name}
                            </p>
                            <p className="truncate text-xs text-white/70">
                              {top.nama_tim ?? "—"}
                            </p>
                          </div>
                        </div>

                        {/* Score badge */}
                        <div className="relative mt-3 flex items-center justify-between">
                          <span className={`rounded-full px-3 py-1 text-sm font-black ${meta.badgeBg}`}>
                            {formatPoints(Number(top.score))} pts
                          </span>
                          <span className="text-lg">
                            {top.kategori === "overall" ? "👑" : top.kategori === "tyfcb" ? "🏅" : "🌟"}
                          </span>
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            );
          })()}

          {/* Active boosters */}
          {activeBoosters.length > 0 ? (
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <p className="text-sm font-black uppercase tracking-[0.1em] text-brand-700">Booster Aktif</p>
                <Link href="/booster" className="text-sm font-bold text-brand-600 hover:underline">
                  Lihat semua Booster →
                </Link>
              </div>
              {activeBoosters.slice(0, 2).map((booster) => (
                <Link
                  key={booster.id}
                  href={`/booster/${booster.id}`}
                  className="group flex items-start gap-4 overflow-hidden rounded-2xl bg-gradient-to-br from-brand-600 via-red-600 to-orange-500 p-4 text-white shadow-lift transition hover:brightness-110 active:scale-[0.99]"
                >
                  <span className="grid h-11 w-11 shrink-0 place-items-center rounded-full bg-white/20">
                    <Zap className="h-5 w-5 fill-white" />
                  </span>
                  <div className="min-w-0 flex-1">
                    <p className="text-[0.68rem] font-bold uppercase tracking-[0.14em] text-white/70">
                      +{Number(booster.poin)} pts · {booster.tanggal_mulai} — {booster.tanggal_berakhir}
                    </p>
                    <p className="mt-0.5 truncate text-lg font-black">{booster.judul}</p>
                    {booster.deskripsi && (
                      <p className="mt-0.5 line-clamp-1 text-sm text-white/80">{booster.deskripsi}</p>
                    )}
                  </div>
                  <ChevronRight className="mt-1 h-5 w-5 shrink-0 text-white/60 transition group-hover:translate-x-0.5 group-hover:text-white" />
                </Link>
              ))}
            </div>
          ) : (
            <div className="flex items-center gap-4 rounded-2xl border border-brand-100 bg-white p-5">
              <span className="grid h-12 w-12 shrink-0 place-items-center rounded-full bg-brand-50 text-brand-600">
                <Zap className="h-6 w-6" />
              </span>
              <div>
                <p className="font-black text-ink">Belum ada booster point minggu ini</p>
                <p className="text-sm text-muted">Grow Coordinator belum set event untuk minggu ini.</p>
              </div>
            </div>
          )}

          {/* Recent TYFCB entries */}
          <section className="glass-panel rounded-2xl p-4">
            <div className="mb-4 flex items-center justify-between">
              <h2 className="text-xl font-black text-ink">TYFCB Terakhir</h2>
              <Link href="/history" className="text-sm font-bold text-brand-600">
                Lihat semua
              </Link>
            </div>
            {recentTyfcb.length === 0 ? (
              <p className="py-6 text-center text-sm text-muted">
                Belum ada TYFCB.{" "}
                <Link href="/submit?type=tyfcb" className="font-bold text-brand-600">
                  Submit sekarang →
                </Link>
              </p>
            ) : (
              <div className="space-y-3">
                {recentTyfcb.map((entry) => (
                  <div
                    className="flex items-center justify-between gap-3 border-b border-brand-100 pb-3 last:border-0 last:pb-0"
                    key={entry.id}
                  >
                    <div className="min-w-0">
                      <p className="truncate font-bold text-ink">
                        Pembeli: {entry.buyer_name ?? "—"}
                      </p>
                      <p className="text-xs text-muted">
                        Rp {Number(entry.nilai).toLocaleString("id-ID")} · {entry.tanggal}
                      </p>
                    </div>
                    <span className={`shrink-0 rounded-full px-3 py-1 text-xs font-bold ${
                      entry.status === "verified"
                        ? "bg-green-50 text-green-700"
                        : entry.status === "rejected"
                          ? "bg-red-50 text-red-700"
                          : "bg-brand-50 text-brand-700"
                    }`}>
                      {entry.status === "verified" && entry.computed_score
                        ? `+${entry.computed_score} pts`
                        : entry.status}
                    </span>
                  </div>
                ))}
              </div>
            )}
          </section>
        </section>

        {/* Sidebar — team standings */}
        <aside className="space-y-6">
          <section className="glass-panel rounded-2xl p-4">
            <h2 className="text-xl font-black text-ink">Team Standings</h2>
            <div className="mt-4 space-y-3">
              {teams.map((team, index) => (
                <div
                  className={`rounded-xl border p-3 ${
                    team.team_id === member.team_id
                      ? "border-brand-300 bg-brand-50"
                      : "border-brand-100 bg-white"
                  }`}
                  key={team.team_id}
                >
                  <div className="flex items-center justify-between gap-3">
                    <div className="flex items-center gap-3">
                      <span className="grid h-9 w-9 place-items-center rounded-full bg-brand-600 text-sm font-black text-white">
                        {index + 1}
                      </span>
                      <div>
                        <p className="font-black text-ink">{team.nama_tim}</p>
                        <p className="text-xs text-muted">
                          TYFCB {Number(team.count_tyfcb)}× · Visitor {Number(team.count_visitor)}
                        </p>
                      </div>
                    </div>
                    <p className="text-lg font-black text-brand-600">
                      {formatPoints(Number(team.score_overall))}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </section>
        </aside>
      </div>
    </AppShell>
  );
}
