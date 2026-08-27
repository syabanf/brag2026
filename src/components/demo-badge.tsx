"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import { FlaskConical, LogOut } from "lucide-react";
import { DEMO_ROLES, type DemoRole } from "@/lib/demo";

const LABELS: Record<DemoRole, string> = {
  admin: "Admin",
  captain: "Captain",
  member: "Member"
};

export function DemoBadge({ role }: { role: DemoRole }) {
  const router = useRouter();
  const [pending, startTransition] = useTransition();
  const [current, setCurrent] = useState<DemoRole>(role);

  async function switchRole(next: DemoRole) {
    setCurrent(next);
    await fetch("/api/demo/role", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ role: next })
    });
    startTransition(() => router.refresh());
  }

  async function exitDemo() {
    await fetch("/api/demo/session", { method: "DELETE" });
    router.replace("/welcome");
    router.refresh();
  }

  return (
    <div className="mb-4 flex items-center gap-2 rounded-xl border border-amber-200/80 bg-amber-50/70 px-2 py-1.5">
      <span
        title="Mode demo — data contoh"
        className="flex shrink-0 items-center gap-1.5 pl-0.5 text-[0.6rem] font-black uppercase tracking-[0.12em] text-amber-700"
      >
        <FlaskConical className="h-3.5 w-3.5" aria-hidden />
        <span className="hidden sm:inline">Demo</span>
      </span>

      <div
        role="group"
        aria-label="Ganti peran demo"
        className="flex min-w-0 flex-1 items-center justify-center gap-0.5 rounded-full bg-white/80 p-0.5"
      >
        {DEMO_ROLES.map((r) => (
          <button
            key={r}
            type="button"
            disabled={pending}
            aria-pressed={current === r}
            onClick={() => switchRole(r)}
            className={`min-w-0 flex-1 truncate rounded-full px-2 py-1.5 text-[0.7rem] font-bold transition disabled:opacity-60 ${
              current === r
                ? "bg-amber-600 text-white"
                : "text-amber-800 hover:bg-amber-100"
            }`}
          >
            {LABELS[r]}
          </button>
        ))}
      </div>

      <button
        type="button"
        onClick={exitDemo}
        aria-label="Keluar dari mode demo"
        className="flex shrink-0 items-center gap-1 rounded-full px-2 py-1.5 text-[0.7rem] font-bold text-amber-800 transition hover:bg-amber-100"
      >
        <LogOut className="h-3.5 w-3.5" aria-hidden />
        <span className="hidden sm:inline">Keluar</span>
      </button>
    </div>
  );
}
