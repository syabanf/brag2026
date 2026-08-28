import { useState } from "react";
import { Link } from "react-router-dom";
import { Bell, Gift, UserPlus } from "lucide-react";
import { api } from "../lib/api";
import { useApi } from "../lib/use-api";
import { formatCurrency } from "../lib/format";
import type { ActivityItem } from "../lib/types";

/** How many events count as "recent" for the unread dot. */
const RECENT_HOURS = 24;

/**
 * Counted inside the fetch rather than during render: reading the clock while
 * rendering makes the output depend on when React happens to re-run the
 * component, which is exactly the instability the purity rule guards against.
 */
function countRecent(items: ActivityItem[]) {
  const cutoff = Date.now() - RECENT_HOURS * 60 * 60 * 1000;
  return items.filter((item) => new Date(item.created_at).getTime() > cutoff).length;
}

export function NotificationBell() {
  const [open, setOpen] = useState(false);
  // Only fetched once the bell is opened: the badge count is derived from the
  // same call, so an unopened bell costs nothing.
  const { data, loading } = useApi(
    () => api.activity(8).then((items) => ({ items, recent: countRecent(items) })),
    [open ? "open" : "closed"],
  );

  const items = data?.items;
  const recent = data?.recent ?? 0;

  return (
    <div className="relative">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        aria-expanded={open}
        aria-label={recent > 0 ? `Notifikasi, ${recent} baru` : "Notifikasi"}
        className="relative grid h-11 w-11 place-items-center rounded-full border border-brand-100 bg-white text-brand-600 transition hover:bg-brand-50"
      >
        <Bell className="h-[1.15rem] w-[1.15rem]" />
        {recent > 0 && (
          <span
            aria-hidden
            className="num absolute -right-0.5 -top-0.5 grid h-5 min-w-5 place-items-center rounded-full bg-brand-600 px-1 text-[0.6rem] font-black text-white"
          >
            {recent > 9 ? "9+" : recent}
          </span>
        )}
      </button>

      {open && (
        <>
          <div className="fixed inset-0 z-30" aria-hidden onClick={() => setOpen(false)} />

          <div className="absolute right-0 top-full z-40 mt-2 w-80 overflow-hidden rounded-2xl border border-brand-100 bg-white shadow-lift">
            <div className="border-b border-brand-50 px-4 py-3">
              <p className="text-sm font-black text-ink">Aktivitas terbaru</p>
            </div>

            {loading && <p className="px-4 py-6 text-center text-sm text-muted">Memuat…</p>}

            {items && items.length === 0 && (
              <p className="px-4 py-6 text-center text-sm text-muted">Belum ada aktivitas.</p>
            )}

            <ul className="max-h-80 divide-y divide-brand-50 overflow-y-auto">
              {items?.map((item) => (
                <li key={`${item.type}-${item.id}`} className="flex items-start gap-2.5 px-4 py-3">
                  <span
                    aria-hidden
                    className={`mt-0.5 grid h-7 w-7 shrink-0 place-items-center rounded-lg ${
                      item.type === "tyfcb"
                        ? "bg-brand-50 text-brand-600"
                        : "bg-amber-50 text-amber-600"
                    }`}
                  >
                    {item.type === "tyfcb" ? (
                      <Gift className="h-3.5 w-3.5" />
                    ) : (
                      <UserPlus className="h-3.5 w-3.5" />
                    )}
                  </span>

                  <div className="min-w-0">
                    <p className="text-xs leading-snug text-ink">
                      <span className="font-bold">{item.actor_name}</span>
                      {item.type === "tyfcb" ? " → " : " mengundang "}
                      <span className="font-bold">{item.target_name}</span>
                    </p>
                    {item.amount != null && (
                      <p className="num text-[0.68rem] text-muted">{formatCurrency(item.amount)}</p>
                    )}
                  </div>
                </li>
              ))}
            </ul>

            <Link
              to="/activity"
              onClick={() => setOpen(false)}
              className="block border-t border-brand-50 px-4 py-3 text-center text-sm font-bold text-brand-600 transition hover:bg-brand-50"
            >
              Lihat semua aktivitas
            </Link>
          </div>
        </>
      )}
    </div>
  );
}
