"use client";

import { Send, Zap } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { BOOSTER_BAND_PRESETS, boosterScorePreview } from "@/lib/booster";

const today = new Date().toISOString().split("T")[0];

export function NewBoosterForm() {
  const router = useRouter();
  const [form, setForm] = useState({
    judul: "",
    deskripsi: "",
    tanggal_mulai: today,
    tanggal_berakhir: today,
    poin: "0",
    multiplier: "1",
    bandIndex: "0",
  });
  const [saving, setSaving] = useState(false);
  const band = BOOSTER_BAND_PRESETS[Number(form.bandIndex)];
  const preview = boosterScorePreview(Number(form.multiplier), band);
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError("");
    const res = await fetch("/api/admin/booster", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        ...form,
        poin: Number(form.poin),
        multiplier: Number(form.multiplier),
        band_min: band.band_min,
        band_max: band.band_max,
      }),
    });
    setSaving(false);
    if (res.ok) {
      router.push("/admin/booster");
      router.refresh();
    } else {
      const d = await res.json();
      setError(d.error ?? "Gagal menyimpan.");
    }
  }

  return (
    <div className="mx-auto max-w-xl">
      <div className="mb-6 flex items-center gap-3">
        <Link href="/admin/booster" className="text-sm font-bold text-brand-600 hover:underline">
          ← Kembali
        </Link>
      </div>
      <div className="mb-6">
        <h1 className="flex items-center gap-3 text-3xl font-black text-ink">
          <Zap className="h-7 w-7 text-brand-600" />
          Tambah Booster
        </h1>
        <p className="mt-2 text-muted">Buat event booster point baru untuk season BRAG 2026.</p>
      </div>

      <form className="glass-panel space-y-5 rounded-2xl p-5" onSubmit={handleSubmit}>
        <label className="block">
          <span className="mb-2 block text-sm font-black text-ink">
            Judul <span className="text-brand-600">*</span>
          </span>
          <input
            className="w-full rounded-2xl border border-brand-100 bg-white px-4 py-3 outline-none focus:ring-2 focus:ring-brand-500"
            placeholder="Contoh: Double Point Weekend"
            value={form.judul}
            onChange={e => setForm({ ...form, judul: e.target.value })}
            required
          />
        </label>

        <label className="block">
          <span className="mb-2 block text-sm font-black text-ink">Deskripsi</span>
          <textarea
            className="w-full rounded-2xl border border-brand-100 bg-white px-4 py-3 outline-none focus:ring-2 focus:ring-brand-500"
            placeholder="Jelaskan mekanisme booster ini…"
            rows={3}
            value={form.deskripsi}
            onChange={e => setForm({ ...form, deskripsi: e.target.value })}
          />
        </label>

        <div className="grid grid-cols-2 gap-4">
          <label className="block">
            <span className="mb-2 block text-sm font-black text-ink">
              Tanggal Mulai <span className="text-brand-600">*</span>
            </span>
            <input
              type="date"
              className="w-full rounded-2xl border border-brand-100 bg-white px-4 py-3 outline-none focus:ring-2 focus:ring-brand-500"
              value={form.tanggal_mulai}
              onChange={e => setForm({ ...form, tanggal_mulai: e.target.value })}
              required
            />
          </label>
          <label className="block">
            <span className="mb-2 block text-sm font-black text-ink">
              Tanggal Berakhir <span className="text-brand-600">*</span>
            </span>
            <input
              type="date"
              className="w-full rounded-2xl border border-brand-100 bg-white px-4 py-3 outline-none focus:ring-2 focus:ring-brand-500"
              value={form.tanggal_berakhir}
              min={form.tanggal_mulai}
              onChange={e => setForm({ ...form, tanggal_berakhir: e.target.value })}
              required
            />
          </label>
        </div>

        <div className="space-y-5 rounded-2xl border border-brand-100 bg-brand-50/40 p-4">
          <p className="text-sm font-black text-ink">Pengali Poin TYFCB</p>

          <div className="grid grid-cols-2 gap-4">
            <label className="block">
              <span className="mb-2 block text-sm font-black text-ink">Multiplier</span>
              <div className="relative">
                <input
                  type="number"
                  min="1"
                  max="10"
                  step="0.5"
                  className="w-full rounded-2xl border border-brand-100 bg-white px-4 py-3 pr-10 outline-none focus:ring-2 focus:ring-brand-500"
                  value={form.multiplier}
                  onChange={e => setForm({ ...form, multiplier: e.target.value })}
                />
                <span className="pointer-events-none absolute inset-y-0 right-4 flex items-center text-sm font-bold text-muted">x</span>
              </div>
            </label>
            <label className="block">
              <span className="mb-2 block text-sm font-black text-ink">Berlaku untuk band</span>
              <select
                className="w-full rounded-2xl border border-brand-100 bg-white px-4 py-3 outline-none focus:ring-2 focus:ring-brand-500"
                value={form.bandIndex}
                onChange={e => setForm({ ...form, bandIndex: e.target.value })}
              >
                {BOOSTER_BAND_PRESETS.map((b, i) => (
                  <option key={b.label} value={i}>{b.label}</option>
                ))}
              </select>
            </label>
          </div>

          <p className="text-xs font-bold text-muted">
            {Number(form.multiplier) > 1
              ? preview
                ? `TYFCB di band ini otomatis bernilai ${preview} selama periode booster.`
                : "Semua band TYFCB dikalikan otomatis selama periode booster."
              : "Multiplier 1x berarti booster ini hanya pengumuman — poin tidak dikalikan otomatis."}
          </p>
        </div>

        <label className="block">
          <span className="mb-2 block text-sm font-black text-ink">Point Bonus Flat</span>
          <div className="relative">
            <input
              type="number"
              min="0"
              className="w-full rounded-2xl border border-brand-100 bg-white px-4 py-3 pr-16 outline-none focus:ring-2 focus:ring-brand-500"
              placeholder="0"
              value={form.poin}
              onChange={e => setForm({ ...form, poin: e.target.value })}
            />
            <span className="pointer-events-none absolute inset-y-0 right-4 flex items-center text-sm font-bold text-muted">pts</span>
          </div>
          <span className="mt-2 block text-xs font-bold text-muted">
            Hanya label pengumuman — poin ini tidak dihitung otomatis, admin memberikannya manual.
          </span>
        </label>

        {error && (
          <p className="rounded-xl bg-red-50 px-4 py-3 text-sm font-bold text-red-700">{error}</p>
        )}

        <Button className="w-full" type="submit" disabled={saving}>
          <Send className="h-5 w-5" />
          {saving ? "Menyimpan…" : "Simpan Booster"}
        </Button>
      </form>
    </div>
  );
}
