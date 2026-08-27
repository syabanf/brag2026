"use client";

import { CalendarDays, Edit2, Plus, Power, Trash2, Zap } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { Button } from "@/components/ui/button";

export type BoosterRow = {
  id: string;
  judul: string;
  deskripsi: string | null;
  tanggal_mulai: string;
  tanggal_berakhir: string;
  poin: number;
  status: "aktif" | "nonaktif";
};

function EditModal({
  booster,
  onClose,
  onSaved,
}: {
  booster: BoosterRow;
  onClose: () => void;
  onSaved: () => void;
}) {
  const [form, setForm] = useState({
    judul: booster.judul,
    deskripsi: booster.deskripsi ?? "",
    tanggal_mulai: booster.tanggal_mulai,
    tanggal_berakhir: booster.tanggal_berakhir,
    poin: String(booster.poin),
  });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  async function handleSave() {
    setSaving(true);
    setError("");
    const res = await fetch(`/api/admin/booster/${booster.id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ ...form, poin: Number(form.poin) }),
    });
    setSaving(false);
    if (res.ok) { onSaved(); onClose(); }
    else { const d = await res.json(); setError(d.error ?? "Gagal menyimpan."); }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm">
      <div className="w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl">
        <h2 className="mb-4 text-xl font-black text-ink">Edit Booster</h2>
        <div className="space-y-4">
          <label className="block">
            <span className="mb-1 block text-sm font-black text-ink">Judul <span className="text-brand-600">*</span></span>
            <input className="w-full rounded-xl border border-brand-100 px-3 py-2.5 outline-none focus:ring-2 focus:ring-brand-500"
              value={form.judul} onChange={e => setForm({ ...form, judul: e.target.value })} />
          </label>
          <label className="block">
            <span className="mb-1 block text-sm font-black text-ink">Deskripsi</span>
            <textarea className="w-full rounded-xl border border-brand-100 px-3 py-2.5 outline-none focus:ring-2 focus:ring-brand-500"
              rows={2} value={form.deskripsi} onChange={e => setForm({ ...form, deskripsi: e.target.value })} />
          </label>
          <div className="grid grid-cols-2 gap-3">
            <label className="block">
              <span className="mb-1 block text-sm font-black text-ink">Tanggal Mulai <span className="text-brand-600">*</span></span>
              <input type="date" className="w-full rounded-xl border border-brand-100 px-3 py-2.5 outline-none focus:ring-2 focus:ring-brand-500"
                value={form.tanggal_mulai} onChange={e => setForm({ ...form, tanggal_mulai: e.target.value })} />
            </label>
            <label className="block">
              <span className="mb-1 block text-sm font-black text-ink">Tanggal Berakhir <span className="text-brand-600">*</span></span>
              <input type="date" className="w-full rounded-xl border border-brand-100 px-3 py-2.5 outline-none focus:ring-2 focus:ring-brand-500"
                value={form.tanggal_berakhir} onChange={e => setForm({ ...form, tanggal_berakhir: e.target.value })} />
            </label>
          </div>
          <label className="block">
            <span className="mb-1 block text-sm font-black text-ink">Point Booster <span className="text-brand-600">*</span></span>
            <input type="number" min="0" className="w-full rounded-xl border border-brand-100 px-3 py-2.5 outline-none focus:ring-2 focus:ring-brand-500"
              value={form.poin} onChange={e => setForm({ ...form, poin: e.target.value })} />
          </label>
        </div>
        {error && <p className="mt-3 rounded-xl bg-red-50 px-3 py-2 text-sm font-bold text-red-700">{error}</p>}
        <div className="mt-5 flex gap-2">
          <Button className="flex-1" onClick={handleSave} disabled={saving}>
            {saving ? "Menyimpan…" : "Simpan"}
          </Button>
          <Button variant="secondary" onClick={onClose}>Batal</Button>
        </div>
      </div>
    </div>
  );
}

export function BoosterClient({ initial }: { initial: BoosterRow[] }) {
  const router = useRouter();
  const [boosters, setBoosters] = useState(initial);
  const [editing, setEditing] = useState<BoosterRow | null>(null);
  const [deleting, setDeleting] = useState<string | null>(null);

  async function reload() {
    const res = await fetch("/api/admin/booster");
    const d = await res.json();
    setBoosters(d.boosters ?? []);
    router.refresh();
  }

  async function toggleStatus(b: BoosterRow) {
    const next = b.status === "aktif" ? "nonaktif" : "aktif";
    await fetch(`/api/admin/booster/${b.id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ status: next }),
    });
    reload();
  }

  async function handleDelete(id: string) {
    setDeleting(id);
    await fetch(`/api/admin/booster/${id}`, { method: "DELETE" });
    setDeleting(null);
    reload();
  }

  return (
    <>
      {editing && (
        <EditModal booster={editing} onClose={() => setEditing(null)} onSaved={reload} />
      )}

      <div className="mb-4 flex items-center justify-between">
        <h2 className="text-xl font-black text-ink">Daftar Booster</h2>
        <Link href="/admin/booster/new">
          <Button>
            <Plus className="h-4 w-4" />
            Tambah Booster
          </Button>
        </Link>
      </div>

      {boosters.length === 0 ? (
        <div className="flex flex-col items-center gap-4 rounded-2xl border border-dashed border-brand-200 bg-white py-16 text-center">
          <span className="grid h-14 w-14 place-items-center rounded-full bg-brand-50 text-brand-600">
            <Zap className="h-7 w-7" />
          </span>
          <div>
            <p className="font-black text-ink">Belum ada booster</p>
            <p className="mt-1 text-sm text-muted">Klik &quot;Tambah Booster&quot; untuk membuat event booster baru.</p>
          </div>
        </div>
      ) : (
        <div className="space-y-3">
          {boosters.map((b) => {
            const isActive = b.status === "aktif";
            return (
              <div
                key={b.id}
                className={`rounded-2xl border p-4 ${isActive ? "border-brand-200 bg-white" : "border-brand-50 bg-brand-50/50 opacity-60"}`}
              >
                <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                  <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <p className="text-lg font-black text-ink">{b.judul}</p>
                      <span className={`rounded-full px-2.5 py-0.5 text-xs font-black ${
                        isActive ? "bg-green-50 text-green-700" : "bg-gray-100 text-gray-500"
                      }`}>
                        {isActive ? "Aktif" : "Nonaktif"}
                      </span>
                      <span className="rounded-full bg-brand-50 px-2.5 py-0.5 text-xs font-black text-brand-700">
                        +{b.poin} pts
                      </span>
                    </div>
                    {b.deskripsi && (
                      <p className="mt-1 text-sm text-muted">{b.deskripsi}</p>
                    )}
                    <div className="mt-2 flex items-center gap-1.5 text-xs text-muted">
                      <CalendarDays className="h-3.5 w-3.5" />
                      {b.tanggal_mulai} — {b.tanggal_berakhir}
                    </div>
                  </div>

                  <div className="flex shrink-0 gap-2">
                    <button
                      title={isActive ? "Nonaktifkan" : "Aktifkan"}
                      onClick={() => toggleStatus(b)}
                      className={`grid h-9 w-9 place-items-center rounded-xl border transition hover:scale-105 ${
                        isActive ? "border-green-200 bg-green-50 text-green-700" : "border-gray-200 bg-gray-50 text-gray-500"
                      }`}
                    >
                      <Power className="h-4 w-4" />
                    </button>
                    <button
                      title="Edit"
                      onClick={() => setEditing(b)}
                      className="grid h-9 w-9 place-items-center rounded-xl border border-brand-100 bg-white text-brand-600 transition hover:bg-brand-50 hover:scale-105"
                    >
                      <Edit2 className="h-4 w-4" />
                    </button>
                    <button
                      title="Hapus"
                      onClick={() => handleDelete(b.id)}
                      disabled={deleting === b.id}
                      className="grid h-9 w-9 place-items-center rounded-xl border border-red-100 bg-white text-red-500 transition hover:bg-red-50 hover:scale-105 disabled:opacity-50"
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </>
  );
}
