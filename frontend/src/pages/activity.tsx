import { Gift, UserPlus } from "lucide-react";
import { api } from "../lib/api";
import { useApi } from "../lib/use-api";
import { formatCurrency } from "../lib/format";
import { Badge, EmptyState, ErrorNote, PageHeader, Spinner } from "../components/ui";

/** Relative time reads better than a date in a feed of recent events. */
function relativeTime(iso: string) {
  const then = new Date(iso).getTime();
  if (Number.isNaN(then)) return "—";

  const seconds = Math.round((Date.now() - then) / 1000);
  if (seconds < 60) return "baru saja";

  const minutes = Math.round(seconds / 60);
  if (minutes < 60) return `${minutes} menit lalu`;

  const hours = Math.round(minutes / 60);
  if (hours < 24) return `${hours} jam lalu`;

  const days = Math.round(hours / 24);
  if (days < 30) return `${days} hari lalu`;

  return new Intl.DateTimeFormat("id-ID", { day: "2-digit", month: "short" }).format(then);
}

export function ActivityPage() {
  const { data, error, loading, reload } = useApi(() => api.activity(100));

  return (
    <div className="space-y-5 lg:space-y-6">
      <PageHeader
        title="Aktivitas Season"
        description="Semua kontribusi yang masuk dari seluruh tim, terbaru di atas."
      />

      {loading && <Spinner />}
      {error && <ErrorNote message={error} onRetry={reload} />}
      {data && data.length === 0 && <EmptyState message="Belum ada aktivitas." />}

      <ul className="space-y-2">
        {data?.map((item) => (
          <li key={`${item.type}-${item.id}`} className="card flex items-start gap-3 p-3.5">
            <span
              aria-hidden
              className={`grid h-9 w-9 shrink-0 place-items-center rounded-xl ${
                item.type === "tyfcb"
                  ? "bg-brand-50 text-brand-600"
                  : "bg-amber-50 text-amber-600"
              }`}
            >
              {item.type === "tyfcb" ? <Gift className="h-4 w-4" /> : <UserPlus className="h-4 w-4" />}
            </span>

            <div className="min-w-0 flex-1">
              <p className="text-sm leading-snug text-ink">
                <span className="font-bold">{item.actor_name}</span>
                {item.type === "tyfcb" ? " mencatat TYFCB dari " : " mengundang "}
                <span className="font-bold">{item.target_name}</span>
              </p>

              <div className="mt-1 flex flex-wrap items-center gap-x-2 gap-y-1">
                <Badge value={item.status} />
                {item.amount != null && (
                  <span className="num text-xs text-muted">{formatCurrency(item.amount)}</span>
                )}
                {item.points != null && (
                  <span className="num text-xs font-bold text-emerald-700">+{item.points} pts</span>
                )}
                <span className="text-xs text-muted">· {relativeTime(item.created_at)}</span>
              </div>
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
}
