import type { LucideIcon } from "lucide-react";
import type { ReactNode } from "react";
import { Loader2 } from "lucide-react";

export type StatTone = "emerald" | "sky" | "amber" | "brand" | "violet" | "orange" | "neutral";

const TONES: Record<StatTone, { chip: string; value: string }> = {
  emerald: { chip: "bg-emerald-50 text-emerald-600", value: "text-emerald-700" },
  sky: { chip: "bg-sky-50 text-sky-600", value: "text-sky-700" },
  amber: { chip: "bg-amber-50 text-amber-600", value: "text-amber-700" },
  brand: { chip: "bg-brand-50 text-brand-600", value: "text-brand-600" },
  violet: { chip: "bg-violet-50 text-violet-600", value: "text-violet-700" },
  orange: { chip: "bg-orange-50 text-orange-600", value: "text-orange-700" },
  neutral: { chip: "bg-brand-50 text-brand-600", value: "text-ink" },
};

export function StatCard({
  icon: Icon,
  label,
  value,
  helper,
  tone = "neutral",
  title,
}: {
  icon: LucideIcon;
  label: string;
  value: string;
  helper: string;
  tone?: StatTone;
  title?: string;
}) {
  const palette = TONES[tone];

  return (
    <div className="card flex flex-col p-3.5 sm:p-4" title={title}>
      <div className="mb-2.5 flex items-start justify-between gap-2">
        <p className="min-w-0 text-[0.62rem] font-bold uppercase leading-tight tracking-[0.1em] text-muted sm:text-[0.68rem]">
          {label}
        </p>
        <span
          aria-hidden
          className={`grid h-7 w-7 shrink-0 place-items-center rounded-lg sm:h-8 sm:w-8 ${palette.chip}`}
        >
          <Icon className="h-[0.9rem] w-[0.9rem] sm:h-4 sm:w-4" />
        </span>
      </div>
      <p className={`num text-lg font-black leading-none sm:text-xl ${palette.value}`}>{value}</p>
      <p className="mt-1.5 text-[0.68rem] font-medium leading-snug text-muted">{helper}</p>
    </div>
  );
}

export function PageHeader({
  eyebrow,
  title,
  description,
  action,
}: {
  eyebrow?: string;
  title: string;
  description?: string;
  action?: ReactNode;
}) {
  return (
    <div className="flex flex-wrap items-start justify-between gap-3">
      <div className="min-w-0">
        {eyebrow && <p className="section-label text-brand-700">{eyebrow}</p>}
        <h1 className="mt-1 text-2xl font-black leading-tight tracking-tight text-ink sm:text-3xl">
          {title}
        </h1>
        {description && (
          <p className="mt-1.5 max-w-2xl text-sm leading-relaxed text-muted">{description}</p>
        )}
      </div>
      {action}
    </div>
  );
}

export function Spinner({ label = "Memuat…" }: { label?: string }) {
  return (
    <div className="flex items-center justify-center gap-2 py-16 text-sm font-semibold text-muted">
      <Loader2 className="h-4 w-4 animate-spin" />
      {label}
    </div>
  );
}

export function ErrorNote({ message, onRetry }: { message: string; onRetry?: () => void }) {
  return (
    <div className="card border-red-200 bg-red-50/60 p-4">
      <p className="text-sm font-semibold text-red-700">{message}</p>
      {onRetry && (
        <button
          type="button"
          onClick={onRetry}
          className="mt-2 text-sm font-bold text-brand-600 underline underline-offset-2"
        >
          Coba lagi
        </button>
      )}
    </div>
  );
}

export function EmptyState({ message }: { message: string }) {
  return (
    <div className="rounded-2xl border border-dashed border-brand-100 bg-white/60 py-14 text-center text-sm text-muted">
      {message}
    </div>
  );
}

const STATUS_STYLES: Record<string, string> = {
  pending: "bg-amber-50 text-amber-700",
  verified: "bg-emerald-50 text-emerald-700",
  rejected: "bg-red-50 text-red-600",
  void: "bg-slate-100 text-slate-500",
  terdaftar: "bg-slate-100 text-slate-600",
  hadir: "bg-sky-50 text-sky-700",
  hadir_penuh: "bg-emerald-50 text-emerald-700",
  merah: "bg-red-50 text-red-600",
  kuning: "bg-amber-50 text-amber-700",
  hijau: "bg-emerald-50 text-emerald-700",
  admin: "bg-brand-50 text-brand-700",
  captain: "bg-amber-50 text-amber-700",
  member: "bg-slate-100 text-slate-500",
};

const STATUS_LABELS: Record<string, string> = {
  hadir_penuh: "Hadir Penuh",
  captain: "Kapten",
};

export function Badge({ value }: { value: string }) {
  const label = STATUS_LABELS[value] ?? value;
  return (
    <span
      className={`inline-block rounded-full px-2.5 py-0.5 text-[0.68rem] font-black capitalize ${
        STATUS_STYLES[value] ?? "bg-slate-100 text-slate-600"
      }`}
    >
      {label}
    </span>
  );
}

export function Tabs<T extends string>({
  tabs,
  active,
  onChange,
}: {
  tabs: { key: T; label: string }[];
  active: T;
  onChange: (key: T) => void;
}) {
  return (
    <div className="no-scrollbar mb-4 flex gap-2 overflow-x-auto">
      {tabs.map((tab) => (
        <button
          key={tab.key}
          type="button"
          onClick={() => onChange(tab.key)}
          aria-pressed={active === tab.key}
          className={`min-h-11 shrink-0 rounded-full px-4 text-sm font-bold transition ${
            active === tab.key
              ? "bg-brand-600 text-white"
              : "border border-brand-100 bg-white text-muted hover:bg-brand-50"
          }`}
        >
          {tab.label}
        </button>
      ))}
    </div>
  );
}
