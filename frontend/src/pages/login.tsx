import { useState, type FormEvent } from "react";
import { Navigate, useNavigate } from "react-router-dom";
import { Eye, EyeOff, Loader2, LogIn, Lock, Mail } from "lucide-react";
import { useAuth } from "../lib/use-auth";
import { ApiError } from "../lib/api";
import { DEMO_ACCOUNTS, DEMO_PASSWORD, demoSignInEnabled, type DemoAccount } from "../lib/demo-accounts";

export function LoginPage() {
  const { user, loading, signIn } = useAuth();
  const navigate = useNavigate();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  // Which demo persona is mid-sign-in, so only that button shows a spinner.
  const [demoBusy, setDemoBusy] = useState<string | null>(null);

  if (!loading && user) {
    return <Navigate to="/" replace />;
  }

  async function enter(withEmail: string, withPassword: string) {
    setError(null);
    try {
      await signIn(withEmail, withPassword);
      navigate("/", { replace: true });
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Terjadi kesalahan.");
    }
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setSubmitting(true);
    try {
      await enter(email, password);
    } finally {
      setSubmitting(false);
    }
  }

  async function signInAs(account: DemoAccount) {
    // Fill the form too: the persona stays visible after a failure, so it is
    // obvious which account was tried.
    setEmail(account.email);
    setPassword(account.password);

    setDemoBusy(account.email);
    try {
      await enter(account.email, account.password);
    } finally {
      setDemoBusy(null);
    }
  }

  return (
    <main className="flex min-h-dvh items-center justify-center px-4 py-10">
      <div className="w-full max-w-sm">
        <header className="mb-7 text-center">
          <h1 className="text-4xl font-black text-brand-600 sm:text-5xl">BRAG 2026</h1>
          <p className="mt-1.5 text-[0.68rem] font-bold uppercase tracking-[0.18em] text-brand-700">
            BNI Grow Annual Challenge
          </p>
        </header>

        <form onSubmit={handleSubmit} className="card p-5">
          <h2 className="text-lg font-black text-ink">Masuk</h2>
          <p className="mt-1 text-sm text-muted">
            Gunakan email dan kata sandi dari panitia.
          </p>

          <label className="mt-4 block">
            <span className="section-label mb-1.5 block">Email</span>
            <span className="relative block">
              <Mail
                aria-hidden
                className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-brand-600"
              />
              <input
                type="email"
                autoComplete="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="member@bnigrow.id"
                className="field pl-9"
              />
            </span>
          </label>

          <label className="mt-3 block">
            <span className="section-label mb-1.5 block">Kata sandi</span>
            <span className="relative block">
              <Lock
                aria-hidden
                className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-brand-600"
              />
              <input
                type={showPassword ? "text" : "password"}
                autoComplete="current-password"
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                className="field pl-9 pr-10"
              />
              <button
                type="button"
                onClick={() => setShowPassword((v) => !v)}
                aria-label={showPassword ? "Sembunyikan kata sandi" : "Lihat kata sandi"}
                className="absolute right-2 top-1/2 grid h-8 w-8 -translate-y-1/2 place-items-center rounded-lg text-muted transition hover:text-ink"
              >
                {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
              </button>
            </span>
          </label>

          {error && (
            <p role="alert" className="mt-3 rounded-xl bg-red-50 px-3 py-2 text-sm font-semibold text-red-700">
              {error}
            </p>
          )}

          <button type="submit" disabled={submitting} className="btn-primary mt-5 w-full">
            {submitting ? (
              <>
                <Loader2 className="h-4 w-4 animate-spin" />
                Memproses…
              </>
            ) : (
              <>
                <LogIn className="h-4 w-4" />
                Masuk
              </>
            )}
          </button>
        </form>

        {demoSignInEnabled && (
          <section aria-labelledby="demo-heading" className="card mt-4 p-5">
            <h2 id="demo-heading" className="text-sm font-black text-ink">
              Masuk cepat sebagai demo
            </h2>
            <p className="mt-1 text-xs leading-relaxed text-muted">
              Akun contoh dari data seed, lengkap dengan transaksi, visitor, dan badge.
            </p>

            <div className="mt-3 grid gap-2">
              {DEMO_ACCOUNTS.map((account) => (
                <button
                  key={account.email}
                  type="button"
                  disabled={submitting || demoBusy !== null}
                  onClick={() => signInAs(account)}
                  className="flex items-center gap-3 rounded-2xl border border-brand-100 bg-white px-3 py-2.5 text-left transition hover:border-brand-200 hover:bg-brand-50 disabled:opacity-50"
                >
                  <span className="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-brand-50 text-brand-600">
                    {demoBusy === account.email ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      <account.icon className="h-4 w-4" />
                    )}
                  </span>
                  <span className="min-w-0 flex-1">
                    <span className="block text-sm font-bold text-ink">{account.label}</span>
                    <span className="block truncate text-xs text-muted">{account.blurb}</span>
                  </span>
                </button>
              ))}
            </div>

            <p className="num mt-3 text-center text-[0.68rem] text-muted">
              Semua akun demo memakai kata sandi {DEMO_PASSWORD}
            </p>
          </section>
        )}

        <p className="mt-4 text-center text-xs text-muted">
          Lupa kata sandi? Hubungi Growth Coordinator.
        </p>
      </div>
    </main>
  );
}
