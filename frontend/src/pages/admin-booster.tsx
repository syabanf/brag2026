import { useState, type FormEvent } from "react";
import { Loader2, Pencil, Plus, Trash2, X, Zap } from "lucide-react";
import { api, ApiError } from "../lib/api";
import { useApi } from "../lib/use-api";
import { formatDate, today } from "../lib/format";
import { Badge, EmptyState, ErrorNote, PageHeader, Spinner } from "../components/ui";
import type { BoosterEvent } from "../lib/types";

export function AdminBoosterPage() {
  const { data, error, loading, reload } = useApi(() => api.admin.boosters.list());
  const [editing, setEditing] = useState<BoosterEvent | null>(null);
  const [creating, setCreating] = useState(false);

  return (
    <div className="space-y-5">
      <PageHeader
        eyebrow="Admin Area"
        title="Booster"
        description="Pengumuman booster yang tampil di dashboard member. Ini terpisah dari event mingguan, yang mengubah pengali poin."
        action={
          <button type="button" onClick={() => setCreating(true)} className="btn-primary shrink-0">
            <Plus className="h-4 w-4" />
            Tambah
          </button>
        }
      />

      {loading && <Spinner />}
      {error && <ErrorNote message={error} onRetry={reload} />}
      {data && data.length === 0 && <EmptyState message="Belum ada booster." />}

      <ul className="space-y-2">
        {data?.map((booster) => (
          <li key={booster.id} className="card p-3.5">
            <div className="flex items-start gap-3">
              <span
                className={`grid h-10 w-10 shrink-0 place-items-center rounded-xl ${
                  booster.status === "aktif"
                    ? "bg-brand-600 text-white"
                    : "bg-slate-100 text-slate-400"
                }`}
              >
                <Zap className="h-5 w-5" />
              </span>

              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-black text-ink">{booster.judul}</p>
                {booster.deskripsi && (
                  <p className="mt-0.5 text-xs leading-relaxed text-muted">{booster.deskripsi}</p>
                )}
                <p className="num mt-1 text-xs text-muted">
                  +{booster.poin} pts · {formatDate(booster.tanggal_mulai)} —{" "}
                  {formatDate(booster.tanggal_berakhir)}
                </p>
              </div>

              <div className="flex shrink-0 items-center gap-1.5">
                <Badge value={booster.status === "aktif" ? "verified" : "void"} />
                <button
                  type="button"
                  aria-label={`Ubah ${booster.judul}`}
                  onClick={() => setEditing(booster)}
                  className="grid h-10 w-10 place-items-center rounded-lg border border-brand-100 text-muted transition hover:text-brand-600"
                >
                  <Pencil className="h-3.5 w-3.5" />
                </button>
                <button
                  type="button"
                  aria-label={`Hapus ${booster.judul}`}
                  onClick={async () => {
                    if (!confirm(`Hapus booster "${booster.judul}"?`)) return;
                    try {
                      await api.admin.boosters.remove(booster.id);
                      reload();
                    } catch (err) {
                      alert(err instanceof ApiError ? err.message : "Gagal menghapus.");
                    }
                  }}
                  className="grid h-10 w-10 place-items-center rounded-lg border border-red-100 text-red-500 transition hover:bg-red-50"
                >
                  <Trash2 className="h-3.5 w-3.5" />
                </button>
              </div>
            </div>
          </li>
        ))}
      </ul>

      {(editing || creating) && (
        <BoosterDialog
          booster={editing}
          onClose={() => {
            setEditing(null);
            setCreating(false);
          }}
          onSaved={() => {
            setEditing(null);
            setCreating(false);
            reload();
          }}
        />
      )}
    </div>
  );
}

function BoosterDialog({
  booster,
  onClose,
  onSaved,
}: {
  booster: BoosterEvent | null;
  onClose: () => void;
  onSaved: () => void;
}) {
  const [form, setForm] = useState({
    judul: booster?.judul ?? "",
    deskripsi: booster?.deskripsi ?? "",
    poin: String(booster?.poin ?? 0),
    tanggal_mulai: booster?.tanggal_mulai?.slice(0, 10) ?? today(),
    tanggal_berakhir: booster?.tanggal_berakhir?.slice(0, 10) ?? today(),
    status: booster?.status ?? "aktif",
  });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setSaving(true);

    const body = {
      judul: form.judul,
      deskripsi: form.deskripsi || null,
      poin: Number(form.poin),
      tanggal_mulai: form.tanggal_mulai,
      tanggal_berakhir: form.tanggal_berakhir,
      status: form.status,
    };

    try {
      if (booster) {
        await api.admin.boosters.update(booster.id, body);
      } else {
        await api.admin.boosters.create(body);
      }
      onSaved();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Gagal menyimpan.");
    } finally {
      setSaving(false);
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-end justify-center bg-ink/40 px-3 pb-3 backdrop-blur-sm sm:items-center sm:p-4"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <form
        onSubmit={submit}
        className="max-h-[85vh] w-full max-w-md overflow-y-auto rounded-3xl bg-white p-5 shadow-2xl"
      >
        <div className="mb-4 flex items-center justify-between gap-3">
          <h3 className="text-lg font-black text-ink">
            {booster ? "Ubah Booster" : "Tambah Booster"}
          </h3>
          <button
            type="button"
            onClick={onClose}
            aria-label="Tutup"
            className="grid h-10 w-10 shrink-0 place-items-center rounded-full text-muted transition hover:bg-brand-50"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="space-y-3">
          <label className="block">
            <span className="section-label mb-1.5 block">Judul *</span>
            <input
              required
              value={form.judul}
              onChange={(e) => setForm({ ...form, judul: e.target.value })}
              className="field"
            />
          </label>

          <label className="block">
            <span className="section-label mb-1.5 block">Deskripsi</span>
            <textarea
              rows={3}
              value={form.deskripsi}
              onChange={(e) => setForm({ ...form, deskripsi: e.target.value })}
              className="field resize-y"
            />
          </label>

          <div className="grid grid-cols-2 gap-3">
            <label className="block">
              <span className="section-label mb-1.5 block">Poin</span>
              <input
                type="number"
                value={form.poin}
                onChange={(e) => setForm({ ...form, poin: e.target.value })}
                className="field num"
              />
            </label>

            <label className="block">
              <span className="section-label mb-1.5 block">Status</span>
              <select
                value={form.status}
                onChange={(e) => setForm({ ...form, status: e.target.value })}
                className="field"
              >
                <option value="aktif">Aktif</option>
                <option value="nonaktif">Nonaktif</option>
              </select>
            </label>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <label className="block">
              <span className="section-label mb-1.5 block">Mulai *</span>
              <input
                type="date"
                required
                value={form.tanggal_mulai}
                onChange={(e) => setForm({ ...form, tanggal_mulai: e.target.value })}
                className="field"
              />
            </label>
            <label className="block">
              <span className="section-label mb-1.5 block">Berakhir *</span>
              <input
                type="date"
                required
                value={form.tanggal_berakhir}
                onChange={(e) => setForm({ ...form, tanggal_berakhir: e.target.value })}
                className="field"
              />
            </label>
          </div>
        </div>

        {error && (
          <p role="alert" className="mt-3 rounded-xl bg-red-50 px-3 py-2 text-sm font-semibold text-red-700">
            {error}
          </p>
        )}

        <div className="mt-5 flex gap-2">
          <button type="button" onClick={onClose} className="btn-secondary flex-1">
            Batal
          </button>
          <button type="submit" disabled={saving} className="btn-primary flex-1">
            {saving && <Loader2 className="h-4 w-4 animate-spin" />}
            Simpan
          </button>
        </div>
      </form>
    </div>
  );
}
