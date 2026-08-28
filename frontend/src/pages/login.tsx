import { useState, type FormEvent } from "react";
import { Navigate, useNavigate } from "react-router-dom";
import { Eye, EyeOff, Loader2, LogIn, Lock, Mail } from "lucide-react";
import { useAuth } from "../lib/auth-context";
import { ApiError } from "../lib/api";

export function LoginPage() {
  const { user, loading, signIn } = useAuth();
  const navigate = useNavigate();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  if (!loading && user) {
    return <Navigate to="/" replace />;
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setSubmitting(true);

    try {
      await signIn(email, password);
      navigate("/", { replace: true });
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Terjadi kesalahan.");
    } finally {
      setSubmitting(false);
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

        <p className="mt-4 text-center text-xs text-muted">
          Lupa kata sandi? Hubungi Growth Coordinator.
        </p>
      </div>
    </main>
  );
}
