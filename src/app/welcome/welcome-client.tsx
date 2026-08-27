"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import { ArrowRight, FlaskConical, Loader2, LogIn } from "lucide-react";
import { DEMO_ACCOUNTS, DEMO_ROLES, type DemoRole } from "@/lib/demo";

export function WelcomeChoices({ demoAvailable }: { demoAvailable: boolean }) {
  const router = useRouter();
  const [role, setRole] = useState<DemoRole>("admin");
  const [starting, setStarting] = useState(false);

  async function startDemo() {
    setStarting(true);
    const res = await fetch("/api/demo/session", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ role })
    });

    if (!res.ok) {
      setStarting(false);
      return;
    }

    router.replace("/");
    router.refresh();
  }

  return (
    <div className="space-y-4">
      {demoAvailable && (
        <section className="rounded-3xl border border-brand-100 bg-white p-5 shadow-glass">
          <div className="flex items-center gap-2">
            <span className="flex h-9 w-9 items-center justify-center rounded-full bg-brand-50 text-brand-600">
              <FlaskConical className="h-[1.15rem] w-[1.15rem]" />
            </span>
            <div className="min-w-0">
              <h2 className="text-base font-black leading-tight text-ink">Mode Demo</h2>
              <p className="text-xs text-muted">Data contoh, tanpa perlu login</p>
            </div>
          </div>

          <p className="mt-3 text-sm leading-relaxed text-muted">
            Jelajahi seluruh aplikasi dengan satu musim yang sudah berjalan — 10 tim,
            100 anggota, dan ratusan transaksi. Tidak ada database yang dipasang.
          </p>

          <fieldset className="mt-4">
            <legend className="mb-2 text-[0.68rem] font-bold uppercase tracking-[0.14em] text-brand-700">
              Masuk sebagai
            </legend>
            <div className="grid gap-2">
              {DEMO_ROLES.map((r) => {
                const active = role === r;
                return (
                  <button
                    key={r}
                    type="button"
                    onClick={() => setRole(r)}
                    aria-pressed={active}
                    className={`flex min-h-14 items-center gap-3 rounded-2xl border px-4 text-left transition ${
                      active
                        ? "border-brand-600 bg-brand-50"
                        : "border-brand-100 bg-white hover:bg-brand-50"
                    }`}
                  >
                    <span
                      aria-hidden
                      className={`flex h-5 w-5 shrink-0 items-center justify-center rounded-full border-2 ${
                        active ? "border-brand-600" : "border-brand-100"
                      }`}
                    >
                      {active && <span className="h-2.5 w-2.5 rounded-full bg-brand-600" />}
                    </span>
                    <span className="min-w-0">
                      <span className="block text-sm font-bold text-ink">
                        {DEMO_ACCOUNTS[r].label}
                      </span>
                      <span className="block text-xs leading-snug text-muted">
                        {DEMO_ACCOUNTS[r].blurb}
                      </span>
                    </span>
                  </button>
                );
              })}
            </div>
          </fieldset>

          <button
            type="button"
            onClick={startDemo}
            disabled={starting}
            className="mt-4 flex min-h-[3.25rem] w-full items-center justify-center gap-2 rounded-full bg-brand-600 px-5 py-3.5 text-sm font-bold text-white transition hover:bg-brand-700 active:scale-[0.98] disabled:opacity-70"
          >
            {starting ? (
              <>
                <Loader2 className="h-4 w-4 animate-spin" />
                Menyiapkan data demo…
              </>
            ) : (
              <>
                Mulai demo
                <ArrowRight className="h-4 w-4" />
              </>
            )}
          </button>
        </section>
      )}

      <section className="rounded-3xl border border-brand-100 bg-white p-5 shadow-glass">
        <div className="flex items-center gap-2">
          <span className="flex h-9 w-9 items-center justify-center rounded-full bg-brand-50 text-brand-600">
            <LogIn className="h-[1.15rem] w-[1.15rem]" />
          </span>
          <div className="min-w-0">
            <h2 className="text-base font-black leading-tight text-ink">Masuk dengan akun</h2>
            <p className="text-xs text-muted">Data sebenarnya dari database</p>
          </div>
        </div>

        <p className="mt-3 text-sm leading-relaxed text-muted">
          Gunakan email dan kata sandi yang diberikan panitia untuk masuk ke musim
          yang sedang berjalan.
        </p>

        <button
          type="button"
          onClick={() => router.push("/login")}
          className="mt-4 flex min-h-[3.25rem] w-full items-center justify-center gap-2 rounded-full border border-brand-100 bg-white px-5 py-3.5 text-sm font-bold text-ink transition hover:bg-brand-50 active:scale-[0.98]"
        >
          Ke halaman login
          <ArrowRight className="h-4 w-4" />
        </button>
      </section>
    </div>
  );
}
