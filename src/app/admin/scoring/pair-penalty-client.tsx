"use client";

import { AlertTriangle, ArrowRight, Check, Loader2 } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import type { PairPenaltyImpact, TyfcbEntryStatus } from "@/lib/scoring-settings";
import { formatCurrency } from "@/lib/utils";

export function PairPenaltyClient({
  initialEnabled,
  initialImpact,
}: {
  initialEnabled: boolean;
  initialImpact: PairPenaltyImpact;
}) {
  const router = useRouter();
  const [enabled, setEnabled] = useState(initialEnabled);
  const [impact, setImpact] = useState(initialImpact);
  const [withCorrections, setWithCorrections] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [done, setDone] = useState("");

  const target = !enabled;

  async function apply() {
    setSaving(true);
    setError("");
    setDone("");

    const res = await fetch("/api/admin/scoring/pair-penalty", {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ enabled: target, apply_corrections: withCorrections }),
    });
    const data = await res.json();
    setSaving(false);

    if (!res.ok) {
      setError(data.error ?? "Gagal menyimpan setelan.");
      return;
    }

    setEnabled(target);
    setDone(
      `Setelan disimpan. ${data.corrected_entries} entri dikoreksi` +
        (data.ledger_rows > 0
          ? `, ${data.ledger_rows} di antaranya ditulis ke ledger.`
          : " (belum terverifikasi, tidak menyentuh ledger).")
    );

    const refreshed = await fetch("/api/admin/scoring/pair-penalty");
    if (refreshed.ok) setImpact((await refreshed.json()).impact);
    router.refresh();
  }

  return (
    <div className="space-y-6">
      <section className="glass-panel rounded-2xl p-5">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h2 className="text-xl font-black text-ink">Pair Penalty</h2>
            <p className="mt-1 max-w-xl text-sm text-muted">
              Transaksi TYFCB berulang antara pasangan pembeli→penjual yang sama nilainya menyusut.
              Kalau dinonaktifkan, semua entri dihitung penuh sesuai band.
            </p>
          </div>
          <span
            className={`shrink-0 self-start rounded-full px-4 py-1.5 text-sm font-bold ${
              enabled ? "bg-green-50 text-green-700" : "bg-neutral-100 text-muted"
            }`}
          >
            {enabled ? "Aktif" : "Nonaktif"}
          </span>
        </div>

        <dl className="mt-4 grid grid-cols-3 gap-2 rounded-xl border border-brand-100 bg-white p-3 text-center">
          {[
            { label: "Transaksi ke-1–2", value: "1.0×" },
            { label: "ke-3–5", value: "0.7×" },
            { label: "ke-6+", value: "0.5×" },
          ].map((tier) => (
            <div key={tier.label}>
              <dt className="text-xs text-muted">{tier.label}</dt>
              <dd className={`text-lg font-black ${enabled ? "text-ink" : "text-muted line-through"}`}>
                {tier.value}
              </dd>
            </div>
          ))}
        </dl>
      </section>

      <section className="glass-panel rounded-2xl p-5">
        <h2 className="text-xl font-black text-ink">
          Dampak jika pair penalty {target ? "diaktifkan" : "dinonaktifkan"}
        </h2>

        {impact.rows.length === 0 ? (
          <p className="mt-4 rounded-xl border border-brand-100 bg-white p-4 text-sm text-muted">
            Tidak ada entri yang skornya berubah. Aman untuk diganti kapan saja.
          </p>
        ) : (
          <>
            <div className="mt-4 grid gap-3 sm:grid-cols-4">
              <Stat label="Sudah diverifikasi" value={String(impact.verified_count)} />
              <Stat label="Belum diverifikasi" value={String(impact.unverified_count)} />
              <Stat label="Member terdampak" value={String(impact.affected_members)} />
              <Stat
                label="Perubahan ke leaderboard"
                value={`${impact.ledger_selisih > 0 ? "+" : ""}${impact.ledger_selisih} pts`}
                accent
              />
            </div>

            <div className="mt-4 overflow-x-auto">
              <table className="w-full min-w-[560px] text-left text-sm">
                <thead>
                  <tr className="border-b border-brand-100 text-xs uppercase tracking-wide text-muted">
                    <th className="py-2 pr-3 font-bold">Member</th>
                    <th className="py-2 pr-3 font-bold">Nilai</th>
                    <th className="py-2 pr-3 text-center font-bold">Ke-</th>
                    <th className="py-2 pr-3 font-bold">Status</th>
                    <th className="py-2 pr-3 text-right font-bold">Skor</th>
                    <th className="py-2 text-right font-bold">Selisih</th>
                  </tr>
                </thead>
                <tbody>
                  {impact.rows.map((row) => (
                    <tr className="border-b border-brand-50 last:border-0" key={row.entry_id}>
                      <td className="py-2 pr-3 font-semibold text-ink">{row.member_name}</td>
                      <td className="py-2 pr-3 text-muted">{formatCurrency(row.nilai)}</td>
                      <td className="py-2 pr-3 text-center text-muted">{row.pair_ordinal}</td>
                      <td className="py-2 pr-3">
                        <StatusPill status={row.status} />
                      </td>
                      <td className="py-2 pr-3 text-right">
                        <span className="text-muted">{row.skor_lama}</span>
                        <ArrowRight className="mx-1 inline h-3 w-3 text-muted" />
                        <span className="font-black text-ink">{row.skor_baru}</span>
                      </td>
                      <td
                        className={`py-2 text-right font-black ${
                          row.selisih > 0 ? "text-green-700" : "text-red-700"
                        }`}
                      >
                        {row.selisih > 0 ? "+" : ""}
                        {row.selisih}
                      </td>
                    </tr>
                  ))}
                </tbody>
                <tfoot>
                  <tr className="border-t-2 border-brand-100 font-black text-ink">
                    <td className="py-2 pr-3" colSpan={4}>
                      Total
                    </td>
                    <td className="py-2 pr-3 text-right">
                      <span className="text-muted">{impact.total_lama}</span>
                      <ArrowRight className="mx-1 inline h-3 w-3 text-muted" />
                      {impact.total_baru}
                    </td>
                    <td
                      className={`py-2 text-right ${
                        impact.total_selisih > 0 ? "text-green-700" : "text-red-700"
                      }`}
                    >
                      {impact.total_selisih > 0 ? "+" : ""}
                      {impact.total_selisih}
                    </td>
                  </tr>
                </tfoot>
              </table>
            </div>

            <label className="mt-4 flex cursor-pointer items-start gap-2.5 rounded-xl border border-brand-100 bg-white p-3">
              <input
                type="checkbox"
                checked={withCorrections}
                onChange={(e) => setWithCorrections(e.target.checked)}
                className="mt-0.5 h-4 w-4 accent-brand-600"
              />
              <span className="text-sm">
                <span className="font-bold text-ink">Sekaligus koreksi skor yang sudah diverifikasi</span>
                <span className="block text-muted">
                  Menulis baris koreksi ke ledger (append-only) supaya leaderboard ikut menyesuaikan.
                  Kalau tidak dicentang, skor yang sudah diverifikasi dibiarkan apa adanya.
                </span>
              </span>
            </label>

            <p className="mt-2 rounded-xl border border-brand-100 bg-white p-3 text-sm text-muted">
              Entri yang <span className="font-bold text-ink">belum diverifikasi</span> selalu
              dihitung ulang mengikuti setelan di atas — poinnya belum masuk ledger, jadi koreksinya
              tidak memengaruhi leaderboard.
            </p>
          </>
        )}

        {error && (
          <p className="mt-4 flex items-center gap-2 rounded-xl bg-red-50 p-3 text-sm font-semibold text-red-700">
            <AlertTriangle className="h-4 w-4 shrink-0" />
            {error}
          </p>
        )}
        {done && (
          <p className="mt-4 flex items-center gap-2 rounded-xl bg-green-50 p-3 text-sm font-semibold text-green-700">
            <Check className="h-4 w-4 shrink-0" />
            {done}
          </p>
        )}

        <div className="mt-4 flex justify-end">
          <Button onClick={apply} disabled={saving}>
            {saving && <Loader2 className="h-4 w-4 animate-spin" />}
            {target ? "Aktifkan pair penalty" : "Nonaktifkan pair penalty"}
          </Button>
        </div>
      </section>
    </div>
  );
}

const STATUS_STYLE: Record<TyfcbEntryStatus, { label: string; className: string }> = {
  verified: { label: "Terverifikasi", className: "bg-green-50 text-green-700" },
  pending: { label: "Pending", className: "bg-amber-50 text-amber-700" },
  rejected: { label: "Ditolak", className: "bg-neutral-100 text-muted" },
};

function StatusPill({ status }: { status: TyfcbEntryStatus }) {
  const style = STATUS_STYLE[status];
  return (
    <span className={`rounded-full px-2 py-0.5 text-xs font-bold ${style.className}`}>
      {style.label}
    </span>
  );
}

function Stat({ label, value, accent }: { label: string; value: string; accent?: boolean }) {
  return (
    <div className="rounded-xl border border-brand-100 bg-white p-3">
      <p className="text-xs text-muted">{label}</p>
      <p className={`mt-0.5 text-2xl font-black ${accent ? "text-brand-600" : "text-ink"}`}>{value}</p>
    </div>
  );
}
