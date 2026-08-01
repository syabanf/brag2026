import type { LucideIcon } from "lucide-react";
import { cn } from "@/lib/utils";

type StatCardProps = {
  icon: LucideIcon;
  label: string;
  value: string;
  helper: string;
  className?: string;
  gradient?: string;
};

export function StatCard({
  icon: Icon,
  label,
  value,
  helper,
  className,
  gradient,
}: StatCardProps) {
  return (
    <div
      className={cn(
        "rounded-2xl p-4",
        gradient
          ? `border ${gradient}`
          : "glass-panel",
        className
      )}
    >
      <div className="mb-3 flex items-center justify-between gap-2">
        <p className={cn("text-[0.7rem] font-semibold uppercase tracking-[0.08em]", gradient ? "text-current opacity-70" : "text-muted")}>
          {label}
        </p>
        <span className={cn("grid h-8 w-8 shrink-0 place-items-center rounded-full", gradient ? "bg-white/40" : "bg-brand-50 text-brand-600")}>
          <Icon className="h-4 w-4" />
        </span>
      </div>
      <p className={cn("text-xl font-black tracking-tight", gradient ? "text-current" : "text-brand-600")}>{value}</p>
      <p className={cn("mt-0.5 text-[0.7rem] font-medium", gradient ? "text-current opacity-70" : "text-muted")}>{helper}</p>
    </div>
  );
}
