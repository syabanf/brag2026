import { useEffect, useState, type ReactNode } from "react";
import { ChevronLeft, ChevronRight, Search, X } from "lucide-react";

/**
 * Debounced search box. Typing fires one request when the user pauses rather
 * than one per keystroke, which matters most on the roster where every query
 * also runs a count.
 */
export function SearchField({
  value,
  onChange,
  placeholder = "Cari…",
}: {
  value: string;
  onChange: (next: string) => void;
  placeholder?: string;
}) {
  // The draft is the source of truth while typing; `value` only seeds it.
  // No sync effect: nothing changes the value from outside, and if a reset
  // button is ever added it should remount this field with a `key` instead.
  const [draft, setDraft] = useState(value);

  useEffect(() => {
    if (draft === value) return;
    const timer = window.setTimeout(() => onChange(draft), 300);
    return () => window.clearTimeout(timer);
  }, [draft, value, onChange]);

  return (
    <span className="relative block min-w-0 flex-1">
      <Search
        aria-hidden
        className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted"
      />
      <input
        value={draft}
        onChange={(e) => setDraft(e.target.value)}
        placeholder={placeholder}
        aria-label={placeholder}
        className="field pl-9 pr-9"
      />
      {draft && (
        <button
          type="button"
          onClick={() => setDraft("")}
          aria-label="Hapus pencarian"
          className="absolute right-2 top-1/2 grid h-7 w-7 -translate-y-1/2 place-items-center rounded-lg text-muted transition hover:text-ink"
        >
          <X className="h-3.5 w-3.5" />
        </button>
      )}
    </span>
  );
}

/** A labelled dropdown that reads as one control with the search box. */
export function FilterSelect({
  label,
  value,
  onChange,
  options,
}: {
  label: string;
  value: string;
  onChange: (next: string) => void;
  options: { value: string; label: string }[];
}) {
  return (
    <select
      aria-label={label}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className="field w-auto shrink-0"
    >
      <option value="">{label}</option>
      {options.map((o) => (
        <option key={o.value} value={o.value}>
          {o.label}
        </option>
      ))}
    </select>
  );
}

export function FilterBar({ children }: { children: ReactNode }) {
  return <div className="mb-4 flex flex-wrap items-center gap-2">{children}</div>;
}

/**
 * Pager over a total. It renders nothing when everything fits on one page —
 * controls that can only be no-ops are noise.
 */
export function Pagination({
  total,
  limit,
  offset,
  onChange,
}: {
  total: number;
  limit: number;
  offset: number;
  onChange: (offset: number) => void;
}) {
  if (total <= limit) return null;

  const page = Math.floor(offset / limit) + 1;
  const pages = Math.ceil(total / limit);
  const from = offset + 1;
  const to = Math.min(offset + limit, total);

  return (
    <nav
      aria-label="Navigasi halaman"
      className="mt-4 flex items-center justify-between gap-3 border-t border-black/[0.06] pt-3"
    >
      <p className="num text-xs text-muted">
        {from}–{to} dari {total}
      </p>

      <div className="flex items-center gap-1.5">
        <button
          type="button"
          onClick={() => onChange(Math.max(0, offset - limit))}
          disabled={page <= 1}
          aria-label="Halaman sebelumnya"
          className="grid h-10 w-10 place-items-center rounded-xl border border-brand-100 bg-white text-muted transition hover:bg-brand-50 hover:text-brand-600 disabled:opacity-40"
        >
          <ChevronLeft className="h-4 w-4" />
        </button>

        <span className="num min-w-16 text-center text-xs font-bold text-ink">
          {page} / {pages}
        </span>

        <button
          type="button"
          onClick={() => onChange(offset + limit)}
          disabled={page >= pages}
          aria-label="Halaman berikutnya"
          className="grid h-10 w-10 place-items-center rounded-xl border border-brand-100 bg-white text-muted transition hover:bg-brand-50 hover:text-brand-600 disabled:opacity-40"
        >
          <ChevronRight className="h-4 w-4" />
        </button>
      </div>
    </nav>
  );
}
