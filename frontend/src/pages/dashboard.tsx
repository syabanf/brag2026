import { Link } from "react-router-dom";
import {
  Banknote,
  ChevronRight,
  Gift,
  Receipt,
  Trophy,
  UserPlus,
  Users,
  Zap,
} from "lucide-react";
import { api } from "../lib/api";
import { useApi } from "../lib/use-api";
import { useAuth } from "../lib/auth-context";
import { formatCurrency, formatCurrencyCompact, formatDate, formatPoints } from "../lib/format";
import { Badge, EmptyState, ErrorNote, PageHeader, Spinner, StatCard } from "../components/ui";

export function DashboardPage() {
  const { user } = useAuth();
  const { data, error, loading, reload } = useApi(() => api.dashboard());

  if (loading) return <Spinner />;
  if (error) return <ErrorNote message={error} onRetry={reload} />;
  if (!data) return null;

  const firstName = user?.full_name.split(" ")[0] ?? "";

  if (!data.member) {
    return (
      <div className="py-20 text-center">
        <p className="text-2xl font-black text-ink">Halo, {firstName}.</p>
        <p className="mt-2 text-muted">Profil member kamu belum terdaftar di season ini.</p>
        <p className="mt-1 text-sm text-muted">
          Hubungi Growth Coordinator untuk mendaftarkan kamu.
        </p>
      </div>
    );
  }

  return (
    <div className="grid gap-5 lg:grid-cols-[1fr_340px]">
      <section className="space-y-5">
        <PageHeader
          eyebrow="Member Dashboard"
          title={`Halo, ${firstName}.`}
          description={`Bantu sesama member closing dan undang tamu.${
            data.my_team ? ` Setiap kontribusi bergerak ke skor ${data.my_team.nama_tim}.` : ""
          }`}
        />

        {/* Row one is season-wide, row two is this member's team. */}
        <div className="grid grid-cols-2 gap-2.5 sm:gap-3 lg:grid-cols-3">
          <StatCard
            icon={Receipt}
            tone="emerald"
            label="Total Transaksi"
            value={`${formatPoints(data.total_tyfcb_tx)}×`}
            helper={`TYFCB verified dari ${data.teams.length} team`}
          />
          <StatCard
            icon={Banknote}
            tone="sky"
            label="Total Nominal TYFCB"
            value={formatCurrencyCompact(data.total_tyfcb_idr)}
            title={formatCurrency(data.total_tyfcb_idr)}
            helper={`Akumulasi semua ${data.teams.length} team`}
          />
          <StatCard
            icon={Users}
            tone="amber"
            label="Total Visitor"
            value={`${formatPoints(data.total_visitor)} tamu`}
            helper={`Akumulasi semua ${data.teams.length} team`}
          />

          {data.my_team && (
            <>
              <StatCard
                icon={Trophy}
                tone="brand"
                label={`Point ${data.my_team.nama_tim}`}
                value={`${formatPoints(data.my_team.score_overall)} pts`}
                helper="Total poin team keseluruhan"
              />
              <StatCard
                icon={Banknote}
                tone="violet"
                label="TYFCB Team"
                value={formatCurrencyCompact(data.my_team.nilai_tyfcb)}
                title={formatCurrency(data.my_team.nilai_tyfcb)}
                helper={`${data.my_team.count_tyfcb}× transaksi verified`}
              />
              <StatCard
                icon={Users}
                tone="orange"
                label="Visitor Team"
                value={`${data.my_team.count_visitor} tamu`}
                helper={`Akumulasi undangan ${data.my_team.nama_tim}`}
              />
            </>
          )}
        </div>

        <div className="grid gap-3 sm:grid-cols-2">
          <QuickAction
            to="/submit?type=tyfcb"
            icon={Gift}
            label="TYFCB"
            helper="Closed business"
          />
          <QuickAction
            to="/submit?type=visitor"
            icon={UserPlus}
            label="Visitor"
            helper="Undang tamu"
          />
        </div>

        <section className="card p-4">
          <div className="mb-3 flex items-center justify-between gap-3">
            <h2 className="text-base font-black text-ink">TYFCB Terakhir</h2>
            <Link to="/history" className="text-sm font-bold text-brand-600">
              Lihat semua
            </Link>
          </div>

          {data.recent_tyfcb.length === 0 ? (
            <EmptyState message="Belum ada transaksi TYFCB." />
          ) : (
            <ul className="divide-y divide-brand-50">
              {data.recent_tyfcb.map((entry) => (
                <li key={entry.id} className="flex items-center justify-between gap-3 py-2.5">
                  <div className="min-w-0">
                    <p className="truncate text-sm font-bold text-ink">
                      Pembeli: {entry.giver_name ?? "—"}
                    </p>
                    <p className="num text-xs text-muted">
                      {formatCurrency(entry.nilai)} · {formatDate(entry.tanggal)}
                    </p>
                  </div>
                  {entry.status === "verified" && entry.computed_score != null ? (
                    <span className="num shrink-0 rounded-full bg-emerald-50 px-2.5 py-0.5 text-[0.68rem] font-black text-emerald-700">
                      +{entry.computed_score} pts
                    </span>
                  ) : (
                    <Badge value={entry.status} />
                  )}
                </li>
              ))}
            </ul>
          )}
        </section>
      </section>

      <aside className="space-y-4">
        <section className="card p-4">
          <h2 className="mb-3 text-base font-black text-ink">Team Standings</h2>
          <ol className="space-y-2">
            {data.teams.slice(0, 10).map((team, index) => {
              const mine = data.my_team?.team_id === team.team_id;
              return (
                <li
                  key={team.team_id}
                  className={`flex items-center gap-2.5 rounded-xl px-2.5 py-2 ${
                    mine ? "bg-brand-50" : ""
                  }`}
                >
                  <span className="num grid h-7 w-7 shrink-0 place-items-center rounded-lg bg-brand-600 text-xs font-black text-white">
                    {index + 1}
                  </span>
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-sm font-black text-ink">{team.nama_tim}</p>
                    <p className="num text-[0.68rem] text-muted">
                      TYFCB {team.count_tyfcb}× · Visitor {team.count_visitor}
                    </p>
                  </div>
                  <span className="num shrink-0 text-sm font-black text-brand-600">
                    {formatPoints(team.score_overall)}
                  </span>
                </li>
              );
            })}
          </ol>
        </section>

        {data.active_boosters.length > 0 && (
          <section className="space-y-2.5">
            <h2 className="section-label text-brand-700">Booster Aktif</h2>
            {data.active_boosters.map((booster) => (
              <Link
                key={booster.id}
                to={`/booster/${booster.id}`}
                className="flex items-center gap-3 rounded-2xl bg-gradient-to-br from-brand-600 to-ember p-4 text-white transition active:scale-[0.99]"
              >
                <span className="grid h-11 w-11 shrink-0 place-items-center rounded-full bg-white/20">
                  <Zap className="h-5 w-5" />
                </span>
                <div className="min-w-0 flex-1">
                  <p className="num text-[0.62rem] font-bold uppercase tracking-[0.12em] text-white/80">
                    +{booster.poin} pts · {formatDate(booster.tanggal_mulai)}
                  </p>
                  <p className="truncate text-sm font-black">{booster.judul}</p>
                </div>
                <ChevronRight className="h-5 w-5 shrink-0 text-white/80" />
              </Link>
            ))}
          </section>
        )}

        {data.badges.length > 0 && (
          <section className="card p-4">
            <h2 className="mb-3 text-base font-black text-ink">Badge Kamu</h2>
            <div className="flex flex-wrap gap-2">
              {data.badges.map((badge) => (
                <span
                  key={badge.badge_code}
                  title={badge.deskripsi}
                  className="flex items-center gap-1.5 rounded-full bg-brand-50 px-3 py-1.5 text-xs font-bold text-brand-700"
                >
                  <span aria-hidden>{badge.ikon ?? "🏅"}</span>
                  {badge.nama}
                </span>
              ))}
            </div>
          </section>
        )}
      </aside>
    </div>
  );
}

function QuickAction({
  to,
  icon: Icon,
  label,
  helper,
}: {
  to: string;
  icon: typeof Gift;
  label: string;
  helper: string;
}) {
  return (
    <Link to={to} className="card flex items-center gap-3 p-4 transition active:scale-[0.99]">
      <span className="grid h-12 w-12 shrink-0 place-items-center rounded-full bg-brand-600 text-white">
        <Icon className="h-5 w-5" />
      </span>
      <div className="min-w-0 flex-1">
        <p className="text-base font-black text-ink">{label}</p>
        <p className="text-sm text-muted">{helper}</p>
      </div>
      <ChevronRight className="h-5 w-5 shrink-0 text-muted" />
    </Link>
  );
}
