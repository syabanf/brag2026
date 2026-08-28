import { Link } from "react-router-dom";
import {
  Banknote,
  CalendarDays,
  ChevronRight,
  Gift,
  Layers,
  Receipt,
  Ticket,
  Users,
  Zap,
} from "lucide-react";
import { api } from "../lib/api";
import { useApi } from "../lib/use-api";
import { formatCurrencyCompact, formatPoints } from "../lib/format";
import { ErrorNote, PageHeader, Spinner, StatCard } from "../components/ui";

const sections = [
  { to: "/admin/tyfcb", icon: Receipt, label: "Verifikasi TYFCB", helper: "Approve atau reject submission" },
  { to: "/admin/visitors", icon: Users, label: "Kelola Visitor", helper: "Update milestone kehadiran" },
  { to: "/admin/members", icon: Users, label: "Kelola Member", helper: "Tim, klasifikasi, role, kata sandi" },
  { to: "/admin/teams", icon: Layers, label: "Kelola Tim", helper: "Tim yang berlaga musim ini" },
  { to: "/admin/classifications", icon: Layers, label: "Klasifikasi Bisnis", helper: "Kategori bisnis member" },
  { to: "/admin/booster", icon: Zap, label: "Booster", helper: "Pengumuman booster musim" },
  { to: "/admin/events", icon: CalendarDays, label: "Event & Bonus", helper: "Jadwal event, pass berkala, prize pool" },
];

export function AdminHomePage() {
  const { data, error, loading, reload } = useApi(() => api.dashboard());
  // Only the total is needed, so one row is enough to fetch.
  const { data: pending } = useApi(() => api.admin.tyfcb.list({ status: "pending", limit: 1 }));

  if (loading) return <Spinner />;
  if (error) return <ErrorNote message={error} onRetry={reload} />;

  const pendingCount = pending?.total ?? 0;

  return (
    <div className="space-y-5">
      <PageHeader
        eyebrow="Admin Area"
        title="Panel Growth Coordinator"
        description="Ringkasan musim dan pintu masuk ke semua alat panitia."
      />

      <div className="grid grid-cols-2 gap-2.5 sm:gap-3 md:grid-cols-4">
        <StatCard
          icon={Receipt}
          tone={pendingCount > 0 ? "amber" : "emerald"}
          label="Menunggu Verifikasi"
          value={`${pendingCount}`}
          helper="Submission TYFCB pending"
        />
        <StatCard
          icon={Banknote}
          tone="sky"
          label="Total Nominal TYFCB"
          value={formatCurrencyCompact(data?.total_tyfcb_idr ?? 0)}
          helper={`${data?.total_tyfcb_tx ?? 0}× transaksi verified`}
        />
        <StatCard
          icon={Users}
          tone="orange"
          label="Total Visitor"
          value={`${formatPoints(data?.total_visitor ?? 0)}`}
          helper="Tamu diundang musim ini"
        />
        <StatCard
          icon={Zap}
          tone="brand"
          label="Booster Aktif"
          value={`${data?.active_boosters.length ?? 0}`}
          helper="Berlaku hari ini"
        />
      </div>

      <section className="grid gap-2.5 md:grid-cols-2 xl:grid-cols-3">
        {sections.map(({ to, icon: Icon, label, helper }) => (
          <Link key={to} to={to} className="card flex items-center gap-3 p-4 transition active:scale-[0.99]">
            <span className="grid h-11 w-11 shrink-0 place-items-center rounded-xl bg-brand-50 text-brand-600">
              <Icon className="h-5 w-5" />
            </span>
            <div className="min-w-0 flex-1">
              <p className="truncate text-sm font-black text-ink">{label}</p>
              <p className="truncate text-xs text-muted">{helper}</p>
            </div>
            <ChevronRight className="h-5 w-5 shrink-0 text-muted" />
          </Link>
        ))}
      </section>

      <section className="grid gap-2.5 md:grid-cols-2 xl:grid-cols-3">
        <Link to="/prizes" className="card flex items-center gap-3 p-4 transition active:scale-[0.99]">
          <span className="grid h-11 w-11 shrink-0 place-items-center rounded-xl bg-amber-50 text-amber-600">
            <Gift className="h-5 w-5" />
          </span>
          <div className="min-w-0 flex-1">
            <p className="text-sm font-black text-ink">Prize Pool</p>
            <p className="text-xs text-muted">Hadiah dan donasi member</p>
          </div>
          <ChevronRight className="h-5 w-5 shrink-0 text-muted" />
        </Link>

        <Link to="/activity" className="card flex items-center gap-3 p-4 transition active:scale-[0.99]">
          <span className="grid h-11 w-11 shrink-0 place-items-center rounded-xl bg-emerald-50 text-emerald-600">
            <Ticket className="h-5 w-5" />
          </span>
          <div className="min-w-0 flex-1">
            <p className="text-sm font-black text-ink">Aktivitas Season</p>
            <p className="text-xs text-muted">Feed semua kontribusi</p>
          </div>
          <ChevronRight className="h-5 w-5 shrink-0 text-muted" />
        </Link>
      </section>
    </div>
  );
}
