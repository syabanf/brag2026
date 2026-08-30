import { useState, type FormEvent } from "react";
import {
  BookOpen,
  Check,
  Copy,
  KeyRound,
  Loader2,
  Plus,
  ShieldAlert,
  Trash2,
  X,
} from "lucide-react";
import { api, ApiError } from "../lib/api";
import { useApi } from "../lib/use-api";
import { formatDate } from "../lib/format";
import { toast } from "../lib/toast-store";
import { EmptyState, ErrorNote, PageHeader, Spinner } from "../components/ui";
import type { APIKey, CreatedAPIKey } from "../lib/types";

const API_BASE = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

export function APIKeysPage() {
  const { data, error, loading, reload } = useApi(() => api.apiKeys.list());
  const [creating, setCreating] = useState(false);
  // The new key, held only long enough to show it. Nothing stores it.
  const [issued, setIssued] = useState<CreatedAPIKey | null>(null);
  const [busy, setBusy] = useState<string | null>(null);

  async function revoke(key: APIKey) {
    if (!window.confirm(`Cabut "${key.nama}"? Integrasi yang memakainya langsung berhenti.`)) {
      return;
    }

    setBusy(key.id);
    try {
      await api.apiKeys.revoke(key.id);
      toast.ok("Kunci dicabut.");
      reload();
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Gagal mencabut kunci.");
    } finally {
      setBusy(null);
    }
  }

  return (
    <div className="space-y-5">
      <PageHeader
        eyebrow="Pengaturan"
        title="Kunci API"
        description="Kredensial untuk integrasi. Kunci bertindak atas nama pemiliknya dan mewarisi perannya."
        action={
          <div className="flex shrink-0 items-center gap-2">
            <a
              href={`${API_BASE}/api/docs/`}
              target="_blank"
              rel="noreferrer"
              className="btn-secondary shrink-0"
            >
              <BookOpen className="h-4 w-4" />
              Dokumentasi
            </a>
            <button type="button" onClick={() => setCreating(true)} className="btn-primary shrink-0">
              <Plus className="h-4 w-4" />
              Buat kunci
            </button>
          </div>
        }
      />

      {loading && <Spinner />}
      {error && <ErrorNote message={error} onRetry={reload} />}
      {data && data.length === 0 && (
        <EmptyState message="Belum ada kunci API. Buat satu untuk mulai mengintegrasikan." />
      )}

      {data && data.length > 0 && (
        <ul className="space-y-2">
          {data.map((key) => (
            <KeyRow key={key.id} apiKey={key} busy={busy === key.id} onRevoke={() => revoke(key)} />
          ))}
        </ul>
      )}

      {creating && (
        <CreateDialog
          onClose={() => setCreating(false)}
          onCreated={(created) => {
            setCreating(false);
            setIssued(created);
            reload();
          }}
        />
      )}

      {issued && <IssuedDialog created={issued} onClose={() => setIssued(null)} />}
    </div>
  );
}

/** Which of the three states a key is in, and how to say so. */
function statusOf(key: APIKey) {
  if (key.revoked_at) return { label: "Dicabut", tone: "bg-red-50 text-red-700" };
  if (key.expires_at && new Date(key.expires_at) <= new Date()) {
    return { label: "Kedaluwarsa", tone: "bg-amber-50 text-amber-700" };
  }
  return { label: "Aktif", tone: "bg-emerald-50 text-emerald-700" };
}

function KeyRow({
  apiKey,
  busy,
  onRevoke,
}: {
  apiKey: APIKey;
  busy: boolean;
  onRevoke: () => void;
}) {
  const status = statusOf(apiKey);
  const live = !apiKey.revoked_at;

  return (
    <li className="card p-3.5">
      <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <p className="truncate text-sm font-black text-ink">{apiKey.nama}</p>
            <span className={`rounded-full px-2 py-0.5 text-[0.68rem] font-bold ${status.tone}`}>
              {status.label}
            </span>
            <span className="rounded-full bg-black/[0.05] px-2 py-0.5 text-[0.68rem] font-bold text-muted">
              {apiKey.read_only ? "Hanya baca" : "Baca & tulis"}
            </span>
          </div>

          <p className="num mt-1 text-xs text-muted">
            <span className="font-semibold text-ink">{apiKey.prefix}…</span> · {apiKey.user_name}
          </p>

          <p className="mt-1 text-xs text-muted">
            Dibuat {formatDate(apiKey.created_at)}
            {apiKey.expires_at && ` · berlaku sampai ${formatDate(apiKey.expires_at)}`}
            {/* "Never used" is the useful signal here — it usually means an
                integration was set up and then forgotten. */}
            {apiKey.last_used_at
              ? ` · terakhir dipakai ${formatDate(apiKey.last_used_at)}`
              : " · belum pernah dipakai"}
          </p>
        </div>

        {live && (
          <button
            type="button"
            disabled={busy}
            onClick={onRevoke}
            className="flex min-h-11 shrink-0 items-center gap-2 rounded-xl border border-red-200 bg-red-50 px-3 text-sm font-bold text-red-600 transition hover:bg-red-100 disabled:opacity-50"
          >
            {busy ? <Loader2 className="h-4 w-4 animate-spin" /> : <Trash2 className="h-4 w-4" />}
            Cabut
          </button>
        )}
      </div>
    </li>
  );
}

function CreateDialog({
  onClose,
  onCreated,
}: {
  onClose: () => void;
  onCreated: (created: CreatedAPIKey) => void;
}) {
  const [nama, setNama] = useState("");
  const [readOnly, setReadOnly] = useState(true);
  const [expiresInDays, setExpiresInDays] = useState(90);
  const [busy, setBusy] = useState(false);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    try {
      onCreated(
        await api.apiKeys.create({
          nama: nama.trim(),
          read_only: readOnly,
          expires_in_days: expiresInDays,
        }),
      );
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Gagal membuat kunci.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <Dialog title="Buat kunci API" onClose={onClose}>
      <form onSubmit={submit}>
        <label className="block">
          <span className="section-label mb-1.5 block">Nama</span>
          <input
            value={nama}
            onChange={(e) => setNama(e.target.value)}
            required
            autoFocus
            placeholder="Integrasi dashboard"
            className="field min-h-12"
          />
          <span className="mt-1 block text-xs text-muted">
            Label agar kunci ini mudah dikenali nanti.
          </span>
        </label>

        <fieldset className="mt-4">
          <legend className="section-label mb-1.5">Akses</legend>
          <div className="grid gap-2">
            <Choice
              checked={readOnly}
              onChange={() => setReadOnly(true)}
              title="Hanya baca"
              blurb="Melayani GET saja. Pilihan yang tepat untuk hampir semua integrasi."
            />
            <Choice
              checked={!readOnly}
              onChange={() => setReadOnly(false)}
              title="Baca & tulis"
              blurb="Bisa mengubah data, termasuk memverifikasi poin."
            />
          </div>
        </fieldset>

        {!readOnly && (
          <p className="mt-3 flex gap-2 rounded-xl bg-amber-50 px-3 py-2.5 text-xs leading-relaxed text-amber-800">
            <ShieldAlert className="mt-0.5 h-4 w-4 shrink-0" />
            <span>
              Kunci ini bisa memberi poin, dan ledger bersifat append-only — koreksinya harus
              ditulis manual. Berikan hanya bila memang perlu menulis.
            </span>
          </p>
        )}

        <label className="mt-4 block">
          <span className="section-label mb-1.5 block">Masa berlaku</span>
          <select
            value={expiresInDays}
            onChange={(e) => setExpiresInDays(Number(e.target.value))}
            className="field min-h-12"
          >
            <option value={30}>30 hari</option>
            <option value={90}>90 hari</option>
            <option value={365}>1 tahun</option>
            <option value={0}>Tanpa batas waktu</option>
          </select>
        </label>

        <div className="mt-5 grid grid-cols-2 gap-2">
          <button type="button" onClick={onClose} className="btn-secondary">
            Batal
          </button>
          <button type="submit" disabled={busy || !nama.trim()} className="btn-primary">
            {busy ? <Loader2 className="h-4 w-4 animate-spin" /> : "Buat"}
          </button>
        </div>
      </form>
    </Dialog>
  );
}

function Choice({
  checked,
  onChange,
  title,
  blurb,
}: {
  checked: boolean;
  onChange: () => void;
  title: string;
  blurb: string;
}) {
  return (
    <label
      className={`flex cursor-pointer gap-3 rounded-2xl border px-3 py-2.5 transition ${
        checked ? "border-brand-600 bg-brand-50" : "border-black/[0.07] hover:bg-brand-50/50"
      }`}
    >
      <input
        type="radio"
        checked={checked}
        onChange={onChange}
        className="mt-1 h-4 w-4 shrink-0 accent-brand-600"
      />
      <span className="min-w-0">
        <span className="block text-sm font-bold text-ink">{title}</span>
        <span className="block text-xs leading-relaxed text-muted">{blurb}</span>
      </span>
    </label>
  );
}

/**
 * The secret, shown once. There is no route that can produce it again — only
 * its digest is stored — so this dialog is deliberately hard to dismiss by
 * accident: no click-outside, no Escape.
 */
function IssuedDialog({ created, onClose }: { created: CreatedAPIKey; onClose: () => void }) {
  const [copied, setCopied] = useState(false);

  async function copy() {
    try {
      await navigator.clipboard.writeText(created.key);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 2000);
    } catch {
      // Clipboard access can be refused; the key is on screen to select.
      toast.error("Tidak bisa menyalin otomatis. Salin manual dari kotak di atas.");
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-end justify-center bg-ink/40 px-3 pb-3 backdrop-blur-sm sm:items-center sm:p-4">
      <div className="w-full max-w-md rounded-3xl bg-white p-5 shadow-2xl">
        <div className="flex items-start gap-3">
          <span className="grid h-10 w-10 shrink-0 place-items-center rounded-2xl bg-emerald-50 text-emerald-600">
            <KeyRound className="h-5 w-5" />
          </span>
          <div className="min-w-0">
            <h3 className="text-lg font-black text-ink">Kunci dibuat</h3>
            <p className="text-sm text-muted">{created.record.nama}</p>
          </div>
        </div>

        <p className="mt-4 flex gap-2 rounded-xl bg-amber-50 px-3 py-2.5 text-xs leading-relaxed text-amber-800">
          <ShieldAlert className="mt-0.5 h-4 w-4 shrink-0" />
          <span>
            Salin sekarang. Kunci ini <strong>tidak bisa ditampilkan lagi</strong> — hanya sidik
            jarinya yang disimpan.
          </span>
        </p>

        <div className="mt-3 rounded-xl border border-black/[0.07] bg-black/[0.03] p-3">
          <code className="block break-all font-mono text-xs leading-relaxed text-ink">
            {created.key}
          </code>
        </div>

        <button type="button" onClick={copy} className="btn-secondary mt-3 w-full">
          {copied ? <Check className="h-4 w-4 text-emerald-600" /> : <Copy className="h-4 w-4" />}
          {copied ? "Tersalin" : "Salin kunci"}
        </button>

        <div className="mt-4">
          <p className="section-label mb-1.5">Contoh pemakaian</p>
          <pre className="overflow-x-auto rounded-xl bg-ink px-3 py-2.5 text-[0.7rem] leading-relaxed text-white/90">
            {`curl -H "X-API-Key: ${created.record.prefix}…" \\
  ${API_BASE}/api/leaderboard`}
          </pre>
        </div>

        <button type="button" onClick={onClose} className="btn-primary mt-4 w-full">
          Selesai — sudah saya simpan
        </button>
      </div>
    </div>
  );
}

function Dialog({
  title,
  onClose,
  children,
}: {
  title: string;
  onClose: () => void;
  children: React.ReactNode;
}) {
  return (
    <div
      className="fixed inset-0 z-50 flex items-end justify-center bg-ink/40 px-3 pb-3 backdrop-blur-sm sm:items-center sm:p-4"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div className="w-full max-w-md rounded-3xl bg-white p-5 shadow-2xl">
        <div className="mb-4 flex items-start justify-between gap-3">
          <h3 className="text-lg font-black text-ink">{title}</h3>
          <button
            type="button"
            onClick={onClose}
            aria-label="Tutup"
            className="grid h-10 w-10 shrink-0 place-items-center rounded-full text-muted transition hover:bg-brand-50 hover:text-ink"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
        {children}
      </div>
    </div>
  );
}
