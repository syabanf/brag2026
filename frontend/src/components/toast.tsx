import { useSyncExternalStore } from "react";
import {
  dismissToast,
  getServerToasts,
  getToasts,
  subscribeToasts,
} from "../lib/toast-store";
import { CheckCircle2, X, XCircle } from "lucide-react";

export function Toaster() {
  const items = useSyncExternalStore(subscribeToasts, getToasts, getServerToasts);

  if (items.length === 0) return null;

  return (
    <div
      // Above the bottom nav on a phone, out of the way of the pill nav on a
      // laptop.
      className="pointer-events-none fixed inset-x-0 bottom-24 z-[60] flex flex-col items-center gap-2 px-4 lg:bottom-24"
      role="status"
      aria-live="polite"
    >
      {items.map((item) => (
        <div
          key={item.id}
          className={`pointer-events-auto flex w-full max-w-sm items-start gap-2.5 rounded-2xl border px-3.5 py-3 shadow-lift ${
            item.tone === "ok"
              ? "border-emerald-200 bg-emerald-50 text-emerald-800"
              : "border-red-200 bg-red-50 text-red-800"
          }`}
        >
          {item.tone === "ok" ? (
            <CheckCircle2 aria-hidden className="mt-0.5 h-4 w-4 shrink-0" />
          ) : (
            <XCircle aria-hidden className="mt-0.5 h-4 w-4 shrink-0" />
          )}

          <p className="min-w-0 flex-1 text-sm font-semibold leading-snug">{item.message}</p>

          <button
            type="button"
            onClick={() => dismissToast(item.id)}
            aria-label="Tutup"
            className="-m-1 shrink-0 rounded-lg p-1 opacity-60 transition hover:opacity-100"
          >
            <X className="h-3.5 w-3.5" />
          </button>
        </div>
      ))}
    </div>
  );
}
