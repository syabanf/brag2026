"use client";

import { Briefcase, Edit2, Plus, Save, Search, Trash2, X } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import type { ClassificationRow } from "@/lib/domain/types";

async function readError(res: Response, fallback: string): Promise<string> {
  const body = await res.json().catch(() => null);
  return body?.error ?? fallback;
}

function EditRow({
  row,
  onCancel,
  onSaved,
  onError,
}: {
  row: ClassificationRow;
  onCancel: () => void;
  onSaved: () => void;
  onError: (message: string) => void;
}) {
  const [nama, setNama] = useState(row.nama);
  const [saving, setSaving] = useState(false);

  async function save() {
    const value = nama.trim();
    if (!value) {
      onError("Nama klasifikasi wajib diisi.");
      return;
    }
    if (value === row.nama) {
      onCancel();
      return;
    }

    setSaving(true);
    const res = await fetch(`/api/admin/classifications/${row.id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ nama: value }),
    });
    setSaving(false);

    if (res.ok) onSaved();
    else onError(await readError(res, "Gagal menyimpan."));
  }

  return (
    <div className="flex items-center gap-2">
      <input
        className="flex-1 rounded-xl border border-brand-300 px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-brand-500"
        value={nama}
        onChange={(e) => setNama(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === "Enter") save();
          if (e.key === "Escape") onCancel();
        }}
        maxLength={60}
        autoFocus
      />
      <button
        onClick={save}
        disabled={saving}
        aria-label="Simpan"
        className="grid h-9 w-9 place-items-center rounded-xl bg-brand-600 text-white hover:bg-brand-700 disabled:opacity-50"
      >
        <Save className="h-4 w-4" />
      </button>
      <button
        onClick={onCancel}
        aria-label="Batal"
        className="grid h-9 w-9 place-items-center rounded-xl border border-brand-100 text-muted hover:bg-brand-50"
      >
        <X className="h-4 w-4" />
      </button>
    </div>
  );
}

export function ClassificationsClient({ initial }: { initial: ClassificationRow[] }) {
  const router = useRouter();
  const [rows, setRows]           = useState(initial);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [addName, setAddName]     = useState("");
  const [adding, setAdding]       = useState(false);
  const [showAdd, setShowAdd]     = useState(false);
  const [filter, setFilter]       = useState("");
  const [error, setError]         = useState("");
  const [notice, setNotice]       = useState("");

  async function reload() {
    const res = await fetch("/api/admin/classifications");
    if (res.ok) {
      const d = await res.json();
      setRows(d.classifications ?? []);
    }
    router.refresh();
  }

  async function handleAdd() {
    const value = addName.trim();
    if (!value) return;

    setAdding(true);
    setError("");
    setNotice("");
    const res = await fetch("/api/admin/classifications", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ nama: value }),
    });
    setAdding(false);

    if (res.ok) {
      setAddName("");
      setShowAdd(false);
      setNotice(`Klasifikasi "${value}" ditambahkan.`);
      reload();
    } else {
      setError(await readError(res, "Gagal menyimpan."));
    }
  }

  async function handleDelete(row: ClassificationRow) {
    setError("");
    setNotice("");

    if (row.jumlah_member > 0) {
      setError(
        `"${row.nama}" masih dipakai ${row.jumlah_member} member. Pindahkan member tersebut terlebih dahulu.`
      );
      return;
    }
    if (!confirm(`Hapus klasifikasi "${row.nama}"?`)) return;

    const res = await fetch(`/api/admin/classifications/${row.id}`, { method: "DELETE" });
    if (res.ok) {
      setNotice(`Klasifikasi "${row.nama}" dihapus.`);
      reload();
    } else {
      setError(await readError(res, "Gagal menghapus."));
    }
  }

  const visible = rows.filter((r) =>
    r.nama.toLowerCase().includes(filter.trim().toLowerCase())
  );
  const terpakai = rows.filter((r) => r.jumlah_member > 0).length;

  return (
    <>
      <div className="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 className="text-xl font-black text-ink">Daftar Klasifikasi</h2>
          <p className="text-xs text-muted">
            {rows.length} klasifikasi · {terpakai} sedang dipakai member
          </p>
        </div>
        <Button
          onClick={() => { setShowAdd(true); setError(""); setNotice(""); }}
        >
          <Plus className="h-4 w-4" />
          Tambah Klasifikasi
        </Button>
      </div>

      {error && (
        <p className="mb-4 rounded-xl bg-red-50 px-4 py-3 text-sm font-bold text-red-700">{error}</p>
      )}
      {notice && (
        <p className="mb-4 rounded-xl bg-green-50 px-4 py-3 text-sm font-bold text-green-700">{notice}</p>
      )}

      {showAdd && (
        <div className="mb-4 flex flex-col gap-2 rounded-2xl border border-brand-200 bg-white p-4 sm:flex-row sm:items-center">
          <input
            className="flex-1 rounded-xl border border-brand-100 px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-brand-500"
            placeholder="Nama klasifikasi baru… (mis. Otomotif)"
            value={addName}
            onChange={(e) => setAddName(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") handleAdd();
              if (e.key === "Escape") { setShowAdd(false); setAddName(""); }
            }}
            maxLength={60}
            autoFocus
          />
          <div className="flex gap-2">
            <Button onClick={handleAdd} disabled={adding}>
              {adding ? "Menyimpan…" : "Simpan"}
            </Button>
            <Button
              variant="secondary"
              onClick={() => { setShowAdd(false); setAddName(""); setError(""); }}
            >
              Batal
            </Button>
          </div>
        </div>
      )}

      {rows.length > 8 && (
        <div className="mb-3 flex items-center gap-2 rounded-xl border border-brand-100 bg-white px-3 py-2">
          <Search className="h-4 w-4 shrink-0 text-muted" />
          <input
            className="flex-1 text-sm outline-none"
            placeholder="Cari klasifikasi…"
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
          />
        </div>
      )}

      <div className="space-y-2">
        {visible.length === 0 ? (
          <p className="rounded-2xl border border-brand-100 bg-white p-8 text-center text-sm text-muted">
            {rows.length === 0
              ? "Belum ada klasifikasi. Tambahkan yang pertama."
              : `Tidak ada klasifikasi cocok dengan "${filter}".`}
          </p>
        ) : (
          visible.map((row) => (
            <div key={row.id} className="rounded-2xl border border-brand-100 bg-white p-4">
              {editingId === row.id ? (
                <EditRow
                  row={row}
                  onCancel={() => setEditingId(null)}
                  onSaved={() => { setEditingId(null); setNotice("Perubahan disimpan."); reload(); }}
                  onError={setError}
                />
              ) : (
                <div className="flex items-center justify-between gap-3">
                  <div className="flex min-w-0 items-center gap-3">
                    <span className="grid h-10 w-10 shrink-0 place-items-center rounded-full bg-brand-50 text-brand-700">
                      <Briefcase className="h-5 w-5" />
                    </span>
                    <div className="min-w-0">
                      <p className="truncate font-black text-ink">{row.nama}</p>
                      <p className="text-xs text-muted">
                        {row.jumlah_member === 0
                          ? "Belum dipakai member"
                          : `Dipakai ${row.jumlah_member} member`}
                      </p>
                    </div>
                  </div>
                  <div className="flex shrink-0 gap-2">
                    <button
                      onClick={() => { setEditingId(row.id); setError(""); setNotice(""); }}
                      aria-label={`Ubah ${row.nama}`}
                      className="grid h-9 w-9 place-items-center rounded-xl border border-brand-100 text-brand-600 hover:bg-brand-50"
                    >
                      <Edit2 className="h-4 w-4" />
                    </button>
                    <button
                      onClick={() => handleDelete(row)}
                      aria-label={`Hapus ${row.nama}`}
                      disabled={row.jumlah_member > 0}
                      title={
                        row.jumlah_member > 0
                          ? "Masih dipakai member — tidak bisa dihapus"
                          : undefined
                      }
                      className="grid h-9 w-9 place-items-center rounded-xl border border-red-100 text-red-500 hover:bg-red-50 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:bg-transparent"
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </div>
                </div>
              )}
            </div>
          ))
        )}
      </div>
    </>
  );
}
