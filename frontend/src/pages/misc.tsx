import { useState, type FormEvent } from "react";
import { Link, useParams } from "react-router-dom";
import { ArrowLeft, Award, KeyRound, Loader2, Zap } from "lucide-react";
import { api, ApiError } from "../lib/api";
import { useApi } from "../lib/use-api";
import { useAuth } from "../lib/use-auth";
import { formatCurrency, formatDate, formatPoints } from "../lib/format";
import { Badge, EmptyState, ErrorNote, PageHeader, Spinner } from "../components/ui";

// ── Booster ───────────────────────────────────────────────────────────────

export function BoosterPage() {
  const { data, error, loading, reload } = useApi(() => api.boosters.list());

  if (loading) return <Spinner />;
  if (error) return <ErrorNote message={error} onRetry={reload} />;

  const boosters = data ?? [];

  return (
    <div className="space-y-5">
      <PageHeader
        title="Booster & Event"
        description="Event mingguan mengubah pengali poin. Booster aktif berlaku untuk semua transaksi yang diverifikasi selama periodenya."
      />

      {boosters.length === 0 ? (
        <EmptyState message="Belum ada booster." />
      ) : (
        <div className="grid gap-3 sm:grid-cols-2">
          {boosters.map((booster) => {
            const active = booster.status === "aktif";
            return (
              <Link
                key={booster.id}
                to={`/booster/${booster.id}`}
                className={`card p-4 transition active:scale-[0.99] ${active ? "" : "opacity-60"}`}
              >
                <div className="flex items-start gap-3">
                  <span
                    className={`grid h-10 w-10 shrink-0 place-items-center rounded-xl ${
                      active ? "bg-brand-600 text-white" : "bg-slate-100 text-slate-400"
                    }`}
                  >
                    <Zap className="h-5 w-5" />
                  </span>
                  <div className="min-w-0 flex-1">
                    <p className="num text-[0.62rem] font-bold uppercase tracking-[0.12em] text-brand-700">
                      +{booster.poin} pts
                    </p>
                    <p className="truncate text-base font-black text-ink">{booster.judul}</p>
                    <p className="mt-0.5 text-xs text-muted">
                      {formatDate(booster.tanggal_mulai)} — {formatDate(booster.tanggal_berakhir)}
                    </p>
                  </div>
                  <Badge value={active ? "verified" : "void"} />
                </div>
              </Link>
            );
          })}
        </div>
      )}
    </div>
  );
}

export function BoosterDetailPage() {
  const { id = "" } = useParams();
  const { data, error, loading } = useApi(() => api.boosters.get(id), [id]);

  if (loading) return <Spinner />;
  if (error) return <ErrorNote message={error} />;
  if (!data) return null;

  return (
    <div className="space-y-5">
      <Link to="/booster" className="inline-flex items-center gap-1.5 text-sm font-bold text-brand-600">
        <ArrowLeft className="h-4 w-4" />
        Semua booster
      </Link>

      <div className="rounded-3xl bg-gradient-to-br from-brand-600 to-ember p-6 text-white">
        <p className="num text-[0.68rem] font-bold uppercase tracking-[0.14em] text-white/80">
          +{data.poin} pts
        </p>
        <h1 className="mt-1 text-2xl font-black">{data.judul}</h1>
        <p className="mt-2 text-sm text-white/90">
          {formatDate(data.tanggal_mulai)} — {formatDate(data.tanggal_berakhir)}
        </p>
      </div>

      {data.deskripsi && (
        <section className="card p-4">
          <h2 className="section-label mb-2">Deskripsi</h2>
          <p className="text-sm leading-relaxed text-muted">{data.deskripsi}</p>
        </section>
      )}
    </div>
  );
}

// ── Awards ────────────────────────────────────────────────────────────────

export function AwardsPage() {
  const { data, error, loading, reload } = useApi(() => api.badges());

  if (loading) return <Spinner />;
  if (error) return <ErrorNote message={error} onRetry={reload} />;

  const badges = data ?? [];

  return (
    <div className="space-y-5">
      <PageHeader
        title="Badge & Penghargaan"
        description="Badge diberikan otomatis saat kamu mencapai milestone tertentu sepanjang musim."
      />

      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
        {badges.map((badge) => (
          <div key={badge.badge_code} className="card flex items-start gap-3 p-4">
            <span
              aria-hidden
              className="grid h-11 w-11 shrink-0 place-items-center rounded-xl bg-brand-50 text-xl"
            >
              {badge.ikon ?? "🏅"}
            </span>
            <div className="min-w-0">
              <p className="text-sm font-black text-ink">{badge.nama}</p>
              <p className="mt-0.5 text-xs leading-relaxed text-muted">{badge.deskripsi}</p>
            </div>
          </div>
        ))}
      </div>

      {badges.length === 0 && <EmptyState message="Belum ada badge terdaftar." />}
    </div>
  );
}

// ── History ───────────────────────────────────────────────────────────────

export function HistoryPage() {
  const { data, error, loading, reload } = useApi(() => api.dashboard());

  if (loading) return <Spinner />;
  if (error) return <ErrorNote message={error} onRetry={reload} />;
  if (!data) return null;

  return (
    <div className="space-y-5">
      <PageHeader
        title="Riwayat Kontribusi"
        description="TYFCB yang tercatat atas namamu, beserta status verifikasinya."
      />

      {data.member_score && (
        <div className="card grid grid-cols-2 gap-3 p-4 sm:grid-cols-4">
          <Figure label="Overall" value={formatPoints(data.member_score.score_overall)} />
          <Figure label="TYFCB" value={formatPoints(data.member_score.score_tyfcb)} />
          <Figure label="Visitor" value={formatPoints(data.member_score.score_visitor)} />
          <Figure label="Bonus" value={formatPoints(data.member_score.score_bonus)} />
        </div>
      )}

      {data.recent_tyfcb.length === 0 ? (
        <EmptyState message="Belum ada transaksi TYFCB." />
      ) : (
        <ul className="space-y-2.5">
          {data.recent_tyfcb.map((entry) => (
            <li key={entry.id} className="card flex items-center justify-between gap-3 p-3.5">
              <div className="min-w-0">
                <p className="truncate text-sm font-bold text-ink">
                  Pembeli: {entry.giver_name ?? "—"}
                </p>
                <p className="num text-xs text-muted">
                  {formatCurrency(entry.nilai)} · {formatDate(entry.tanggal)}
                </p>
              </div>
              <div className="shrink-0 text-right">
                <Badge value={entry.status} />
                {entry.computed_score != null && (
                  <p className="num mt-1 text-xs font-bold text-brand-700">
                    +{entry.computed_score} pts
                  </p>
                )}
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

function Figure({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="section-label">{label}</p>
      <p className="num mt-1 text-xl font-black text-brand-600">{value}</p>
    </div>
  );
}

// ── Profile ───────────────────────────────────────────────────────────────

export function ProfilePage() {
  const { user, member } = useAuth();
  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [result, setResult] = useState<{ message: string; tone: "ok" | "error" } | null>(null);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setResult(null);
    setSubmitting(true);

    try {
      await api.auth.changePassword(current, next);
      // Every session is dropped server-side, so the next request will 401 and
      // bounce the user to the login screen.
      setResult({ message: "Kata sandi diganti. Silakan masuk kembali.", tone: "ok" });
      setCurrent("");
      setNext("");
    } catch (err) {
      setResult({
        message: err instanceof ApiError ? err.message : "Gagal mengganti kata sandi.",
        tone: "error",
      });
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="space-y-5">
      <PageHeader title="Profil" description="Detail akun dan keamanan." />

      <section className="card p-4">
        <dl className="space-y-3">
          <Row label="Nama" value={user?.full_name ?? "—"} />
          <Row label="Email" value={user?.email ?? "—"} />
          <Row label="Role" value={<Badge value={user?.role ?? "member"} />} />
          <Row label="Tim" value={member?.nama_tim ?? "—"} />
          <Row label="Klasifikasi" value={member?.klasifikasi_nama ?? "—"} />
          <Row
            label="Status warna"
            value={member ? <Badge value={member.color_status} /> : "—"}
          />
        </dl>
      </section>

      <form onSubmit={handleSubmit} className="card p-4">
        <h2 className="flex items-center gap-2 text-base font-black text-ink">
          <KeyRound className="h-[1.1rem] w-[1.1rem] text-brand-600" />
          Ganti kata sandi
        </h2>

        <label className="mt-4 block">
          <span className="section-label mb-1.5 block">Kata sandi saat ini</span>
          <input
            type="password"
            required
            autoComplete="current-password"
            value={current}
            onChange={(e) => setCurrent(e.target.value)}
            className="field"
          />
        </label>

        <label className="mt-3 block">
          <span className="section-label mb-1.5 block">Kata sandi baru</span>
          <input
            type="password"
            required
            minLength={6}
            autoComplete="new-password"
            value={next}
            onChange={(e) => setNext(e.target.value)}
            className="field"
          />
          <span className="mt-1 block text-xs text-muted">Minimal 6 karakter.</span>
        </label>

        {result && (
          <p
            role="alert"
            className={`mt-3 rounded-xl px-3 py-2 text-sm font-semibold ${
              result.tone === "ok" ? "bg-emerald-50 text-emerald-700" : "bg-red-50 text-red-700"
            }`}
          >
            {result.message}
          </p>
        )}

        <button type="submit" disabled={submitting} className="btn-primary mt-5 w-full">
          {submitting && <Loader2 className="h-4 w-4 animate-spin" />}
          {submitting ? "Menyimpan…" : "Simpan"}
        </button>
      </form>
    </div>
  );
}

function Row({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between gap-3">
      <dt className="text-sm text-muted">{label}</dt>
      <dd className="text-sm font-bold text-ink">{value}</dd>
    </div>
  );
}

export function NotFoundPage() {
  return (
    <div className="py-24 text-center">
      <Award className="mx-auto h-10 w-10 text-brand-600" />
      <p className="mt-3 text-xl font-black text-ink">Halaman tidak ditemukan</p>
      <Link to="/" className="mt-3 inline-block text-sm font-bold text-brand-600 underline">
        Kembali ke dashboard
      </Link>
    </div>
  );
}
