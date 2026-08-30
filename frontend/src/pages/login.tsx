import { useEffect, useState, type FormEvent } from "react";
import { Navigate, useNavigate } from "react-router-dom";
import { Eye, EyeOff, Loader2, LogIn, Lock, Mail, Trophy, Users, Handshake } from "lucide-react";
import { useAuth } from "../lib/use-auth";
import { api, ApiError } from "../lib/api";
import { formatPoints } from "../lib/format";
import {
  DEMO_ACCOUNTS,
  DEMO_PASSWORD,
  demoSignInEnabled,
  type DemoAccount,
} from "../lib/demo-accounts";
import type { TeamScore } from "../lib/types";

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
    <main className="grid min-h-dvh lg:grid-cols-[1.1fr_1fr]">
      <ShowcasePanel />

      <div className="flex items-center justify-center px-4 py-8 sm:px-8">
        <div className="w-full max-w-sm">
          <form onSubmit={handleSubmit}>
            <h2 className="text-2xl font-black tracking-tight text-ink">Masuk</h2>
            <p className="mt-1 text-sm text-muted">Gunakan email dan kata sandi dari panitia.</p>

            <label className="mt-6 block">
              <span className="section-label mb-1.5 block">Email</span>
              <span className="relative block">
                <Mail
                  aria-hidden
                  className="pointer-events-none absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-brand-600"
                />
                <input
                  type="email"
                  autoComplete="email"
                  required
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="member@bnigrow.id"
                  className="field min-h-12 pl-10"
                />
              </span>
            </label>

            <label className="mt-3 block">
              <span className="section-label mb-1.5 block">Kata sandi</span>
              <span className="relative block">
                <Lock
                  aria-hidden
                  className="pointer-events-none absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-brand-600"
                />
                <input
                  type={showPassword ? "text" : "password"}
                  autoComplete="current-password"
                  required
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="••••••••"
                  className="field min-h-12 pl-10 pr-11"
                />
                <button
                  type="button"
                  onClick={() => setShowPassword((v) => !v)}
                  aria-label={showPassword ? "Sembunyikan kata sandi" : "Lihat kata sandi"}
                  className="absolute right-2 top-1/2 grid h-9 w-9 -translate-y-1/2 place-items-center rounded-lg text-muted transition hover:bg-brand-50 hover:text-ink"
                >
                  {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </button>
              </span>
            </label>

            {error && (
              <p
                role="alert"
                className="mt-4 rounded-xl border border-red-100 bg-red-50 px-3 py-2.5 text-sm font-semibold text-red-700"
              >
                {error}
              </p>
            )}

            <button type="submit" disabled={submitting} className="btn-primary mt-6 w-full">
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
            <section aria-labelledby="demo-heading" className="mt-8">
              {/* A labelled rule rather than another card: nesting a second
                  panel inside the form column made the page read as two
                  competing forms. */}
              <div className="flex items-center gap-3">
                <span className="h-px flex-1 bg-black/[0.07]" />
                <h2 id="demo-heading" className="section-label">
                  atau coba sebagai
                </h2>
                <span className="h-px flex-1 bg-black/[0.07]" />
              </div>

              <div className="mt-4 grid gap-2">
                {DEMO_ACCOUNTS.map((account) => (
                  <button
                    key={account.email}
                    type="button"
                    disabled={submitting || demoBusy !== null}
                    onClick={() => signInAs(account)}
                    className="group flex items-center gap-3 rounded-2xl border border-black/[0.07] bg-white px-3 py-2.5 text-left transition hover:border-brand-200 hover:bg-brand-50 disabled:opacity-50"
                  >
                    <span className="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-brand-50 text-brand-600 transition group-hover:bg-white">
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
                    <LogIn className="h-4 w-4 shrink-0 text-transparent transition group-hover:text-brand-600" />
                  </button>
                ))}
              </div>

              <p className="num mt-3 text-center text-[0.68rem] text-muted">
                Kata sandi semua akun demo: {DEMO_PASSWORD}
              </p>
            </section>
          )}

          <p className="mt-8 text-center text-xs text-muted">
            Lupa kata sandi? Hubungi Growth Coordinator.
          </p>
        </div>
      </div>
    </main>
  );
}

/**
 * The brand half of the page. On a laptop it carries the standings, which is
 * the most honest thing to put in front of someone about to sign in: the
 * competition is already running and here is where it stands. On a phone it
 * collapses to a strip, because a login screen that needs scrolling before
 * reaching the password field is a worse login screen.
 */
function ShowcasePanel() {
  const [teams, setTeams] = useState<TeamScore[] | null>(null);

  useEffect(() => {
    let cancelled = false;
    // Failure is silent: the standings are decoration here, and nobody should
    // be blocked from signing in because a leaderboard query was slow.
    api.leaderboard
      .public()
      .then((data) => {
        if (!cancelled) setTeams(data.teams.slice(0, 4));
      })
      .catch(() => {});
    return () => {
      cancelled = true;
    };
  }, []);

  const totals = teams?.reduce(
    (acc, t) => ({
      tx: acc.tx + t.count_tyfcb,
      guests: acc.guests + t.count_visitor,
    }),
    { tx: 0, guests: 0 },
  );

  return (
    <aside className="relative isolate overflow-hidden bg-brand-600 px-6 py-7 text-white lg:flex lg:flex-col lg:justify-between lg:px-12 lg:py-14">
      {/* Two soft washes give the flat brand red some depth without an image
          to download. */}
      <span
        aria-hidden
        className="pointer-events-none absolute -right-24 -top-24 -z-10 h-80 w-80 rounded-full bg-white/10 blur-3xl"
      />
      <span
        aria-hidden
        className="pointer-events-none absolute -bottom-32 -left-20 -z-10 h-96 w-96 rounded-full bg-ember/25 blur-3xl"
      />

      <div>
        <p className="text-[0.68rem] font-bold uppercase tracking-[0.2em] text-white/70">
          BNI Grow Annual Challenge
        </p>
        <h1 className="mt-1.5 text-3xl font-black tracking-tight lg:mt-2 lg:text-6xl">BRAG 2026</h1>
        <p className="mt-2 max-w-md text-sm leading-relaxed text-white/80 lg:mt-3 lg:text-base">
          Catat TYFCB, undang visitor, kumpulkan badge. Setiap kontribusi menggerakkan skor tim
          sepanjang dua belas minggu.
        </p>
      </div>

      {/* Below the fold on a phone, so the form is what greets you there. */}
      <div className="mt-8 hidden lg:block">
        <div className="flex items-center gap-2 text-white/70">
          <Trophy className="h-4 w-4" />
          <span className="section-label text-white/70">Klasemen sementara</span>
        </div>

        <ol className="mt-4 space-y-2">
          {(teams ?? Array.from({ length: 4 }, () => null)).map((team, i) => (
            <li
              key={team?.team_id ?? i}
              className="flex items-center gap-3 rounded-2xl bg-white/10 px-3.5 py-2.5 backdrop-blur-sm"
            >
              <span className="num grid h-7 w-7 shrink-0 place-items-center rounded-lg bg-white/15 text-xs font-black">
                {i + 1}
              </span>
              {team ? (
                <>
                  <span className="min-w-0 flex-1 truncate text-sm font-bold">{team.nama_tim}</span>
                  <span className="num shrink-0 text-sm font-black">
                    {formatPoints(team.score_overall)}
                  </span>
                </>
              ) : (
                // A pulsing bar rather than a spinner: the row keeps its
                // height, so nothing shifts when the real name arrives.
                <span className="h-4 flex-1 animate-pulse rounded bg-white/15" />
              )}
            </li>
          ))}
        </ol>

        {totals && (
          <dl className="mt-6 flex gap-8">
            <div>
              <dt className="flex items-center gap-1.5 text-xs text-white/70">
                <Handshake className="h-3.5 w-3.5" />
                Transaksi
              </dt>
              <dd className="num mt-0.5 text-2xl font-black">{totals.tx}</dd>
            </div>
            <div>
              <dt className="flex items-center gap-1.5 text-xs text-white/70">
                <Users className="h-3.5 w-3.5" />
                Tamu diundang
              </dt>
              <dd className="num mt-0.5 text-2xl font-black">{totals.guests}</dd>
            </div>
          </dl>
        )}
      </div>
    </aside>
  );
}
