import type { LucideIcon } from "lucide-react";
import { cn } from "@/lib/utils";

export type StatTone =
  | "emerald"
  | "sky"
  | "amber"
  | "brand"
  | "violet"
  | "orange"
  | "neutral";

const TONES: Record<StatTone, { chip: string; value: string }> = {
  emerald: { chip: "bg-emerald-50 text-emerald-600", value: "text-emerald-700" },
  sky:     { chip: "bg-sky-50 text-sky-600",         value: "text-sky-700" },
  amber:   { chip: "bg-amber-50 text-amber-600",     value: "text-amber-700" },
  brand:   { chip: "bg-brand-50 text-brand-600",     value: "text-brand-600" },
  violet:  { chip: "bg-violet-50 text-violet-600",   value: "text-violet-700" },
  orange:  { chip: "bg-orange-50 text-orange-600",   value: "text-orange-700" },
  neutral: { chip: "bg-brand-50 text-brand-600",     value: "text-ink" }
};

type StatCardProps = {
  icon: LucideIcon;
  label: string;
  value: string;
  helper: string;
  className?: string;
  tone?: StatTone;
};

export function StatCard({
  icon: Icon,
  label,
  value,
  helper,
  className,
  tone = "neutral"
}: StatCardProps) {
  const palette = TONES[tone];

  return (
    <div className={cn("card flex flex-col p-3.5 sm:p-4", className)}>
      <div className="mb-2.5 flex items-start justify-between gap-2">
        <p className="min-w-0 text-[0.62rem] font-bold uppercase leading-tight tracking-[0.1em] text-muted sm:text-[0.68rem]">
          {label}
        </p>
        <span
          aria-hidden
          className={cn(
            "grid h-7 w-7 shrink-0 place-items-center rounded-lg sm:h-8 sm:w-8",
            palette.chip
          )}
        >
          <Icon className="h-[0.9rem] w-[0.9rem] sm:h-4 sm:w-4" />
        </span>
      </div>

      <p
        className={cn(
          "num text-lg font-black leading-none sm:text-xl",
          palette.value
        )}
      >
        {value}
      </p>

      <p className="mt-1.5 text-[0.68rem] font-medium leading-snug text-muted">
        {helper}
      </p>
    </div>
  );
}
