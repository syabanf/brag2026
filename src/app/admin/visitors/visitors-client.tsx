"use client";

import { Ban, Pencil, Star, UserCheck, X } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";

export type VisitorStatus = "terdaftar" | "hadir" | "hadir_penuh";

export type VisitorRow = {
  id: string;
  inviter_name: string;
  visitor_name: string;
  kontak: string;
  tanggal_undang: string;
  status_hadir: VisitorStatus;
  is_converted: boolean;
  is_void: boolean;
};

type Filter = "all" | VisitorStatus;

const STATUS_STYLE: Record<VisitorStatus, string> = {
  terdaftar:   "bg-brand-50 text-brand-700",
  hadir:       "bg-blue-50 text-blue-700",
  hadir_penuh: "bg-green-50 text-green-700",
};
const STATUS_LABEL: Record<VisitorStatus, string> = {
  terdaftar:   "Terdaftar",
  hadir:       "Hadir",
  hadir_penuh: "Hadir Penuh",
};
const STATUS_ORDER: VisitorStatus[] = ["terdaftar", "hadir", "hadir_penuh"];

// Mirrors STATUS_CUMULATIVE in the PATCH route — points a visitor has earned
// once it reaches a status, so a change is worth the difference between the two.
const STATUS_CUMULATIVE: Record<VisitorStatus, number> = {
  terdaftar:   0,
  hadir:       20,
  hadir_penuh: 50,
};
const CONVERSION_POINTS = 100;

const NEXT_STATUS: Partial<Record<VisitorStatus, VisitorStatus>> = {
  terdaftar: "hadir",
  hadir:     "hadir_penuh",
};

function formatDelta(points: number): string {
  if (points === 0) return "0 pts";
  return `${points > 0 ? "+" : "−"}${Math.abs(points)} pts`;
}

export function VisitorsAdminClient({ initial }: { initial: VisitorRow[] }) {
  const router = useRouter();
  const [visitors, setVisitors] = useState(initial);
  const [filter, setFilter] = useState<Filter>("all");
  const [loading, setLoading] = useState<string | null>(null);
  const [editing, setEditing] = useState<string | null>(null);
  const [error, setError] = useState("");

  async function patch(id: string, body: object, optimistic: Partial<VisitorRow>) {
    setLoading(id);
    setError("");
    const res = await fetch(`/api/admin/visitors/${id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    const d = await res.json();
    setLoading(null);
    if (!res.ok) { setError(d.error ?? "Gagal."); router.refresh(); return; }
    setVisitors((prev) => prev.map((v) => v.id === id ? { ...v, ...optimistic } : v));
    setEditing(null);
    router.refresh();
  }

  function changeStatus(v: VisitorRow, target: VisitorStatus) {
    const delta = STATUS_CUMULATIVE[target] - STATUS_CUMULATIVE[v.status_hadir];
    if (delta < 0 && !confirm(
      `Ubah status ${v.visitor_name} dari ${STATUS_LABEL[v.status_hadir]} ke ${STATUS_LABEL[target]}?\n` +
      `Poin ${v.inviter_name} akan dikurangi ${Math.abs(delta)}.`
    )) return;
    patch(v.id, { status_hadir: target }, { status_hadir: target });
  }

  function changeConversion(v: VisitorRow, target: boolean) {
    if (!target && !confirm(
      `Batalkan konversi ${v.visitor_name}?\nPoin ${v.inviter_name} akan dikurangi ${CONVERSION_POINTS}.`
    )) return;
    patch(v.id, { is_converted: target }, { is_converted: target });
  }

  const counts = {
    all:         visitors.length,
    terdaftar:   visitors.filter((v) => v.status_hadir === "terdaftar").length,
    hadir:       visitors.filter((v) => v.status_hadir === "hadir").length,
    hadir_penuh: visitors.filter((v) => v.status_hadir === "hadir_penuh").length,
  };

  const visible = filter === "all" ? visitors : visitors.filter((v) => v.status_hadir === filter);

  return (
    <>
      {/* Filter tabs */}
      <div className="mb-4 flex gap-2 overflow-x-auto">
        {(["all", "terdaftar", "hadir", "hadir_penuh"] as Filter[]).map((f) => (
          <button
            key={f}
            onClick={() => setFilter(f)}
            className={`shrink-0 rounded-full px-4 py-1.5 text-sm font-bold transition ${
              filter === f ? "bg-brand-600 text-white" : "bg-white border border-brand-100 text-muted hover:text-ink"
            }`}
          >
            {f === "all" ? "Semua" : STATUS_LABEL[f]} ({counts[f]})
          </button>
        ))}
      </div>

      {error && (
        <p className="mb-4 rounded-xl bg-red-50 px-4 py-3 text-sm font-bold text-red-700">{error}</p>
      )}

      {visible.length === 0 ? (
        <div className="rounded-2xl border border-dashed border-brand-200 bg-white py-16 text-center text-muted">
          Tidak ada visitor.
        </div>
      ) : (
        <div className="space-y-3">
          {visible.map((v) => {
            const nextStatus = NEXT_STATUS[v.status_hadir];
            const isEditing = editing === v.id;
            const busy = loading === v.id;
            return (
              <div key={v.id} className={`rounded-2xl border border-brand-100 bg-white p-4 ${v.is_void ? "opacity-60" : ""}`}>
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <p className="font-black text-ink">{v.visitor_name}</p>
                      <span className={`rounded-full px-2.5 py-0.5 text-xs font-black ${STATUS_STYLE[v.status_hadir]}`}>
                        {STATUS_LABEL[v.status_hadir]}
                      </span>
                      {v.is_converted && (
                        <span className="flex items-center gap-1 rounded-full bg-purple-50 px-2.5 py-0.5 text-xs font-black text-purple-700">
                          <Star className="h-3 w-3" /> Member
                        </span>
                      )}
                      {v.is_void && (
                        <span className="flex items-center gap-1 rounded-full bg-slate-100 px-2.5 py-0.5 text-xs font-black text-slate-600">
                          <Ban className="h-3 w-3" /> Void
                        </span>
                      )}
                    </div>
                    <p className="mt-1 text-sm text-muted">
                      Diundang oleh <span className="font-bold">{v.inviter_name}</span>
                      {" · "}{v.kontak} · {v.tanggal_undang}
                    </p>
                  </div>

                  {/* Action buttons */}
                  {!v.is_void && (
                    <div className="flex shrink-0 flex-wrap gap-2">
                      {nextStatus && (
                        <button
                          onClick={() => changeStatus(v, nextStatus)}
                          disabled={busy}
                          className="flex items-center gap-1.5 rounded-xl bg-brand-600 px-3 py-2 text-xs font-black text-white hover:bg-brand-700 disabled:opacity-50"
                        >
                          <UserCheck className="h-3.5 w-3.5" />
                          {STATUS_LABEL[nextStatus]} ({formatDelta(STATUS_CUMULATIVE[nextStatus] - STATUS_CUMULATIVE[v.status_hadir])})
                        </button>
                      )}
                      {!v.is_converted && (
                        <button
                          onClick={() => changeConversion(v, true)}
                          disabled={busy}
                          className="flex items-center gap-1.5 rounded-xl border border-purple-200 bg-purple-50 px-3 py-2 text-xs font-black text-purple-700 hover:bg-purple-100 disabled:opacity-50"
                        >
                          <Star className="h-3.5 w-3.5" />
                          Konversi (+{CONVERSION_POINTS} pts)
                        </button>
                      )}
                      <button
                        onClick={() => { setEditing(isEditing ? null : v.id); setError(""); }}
                        disabled={busy}
                        className="flex items-center gap-1.5 rounded-xl border border-brand-100 px-3 py-2 text-xs font-black text-muted hover:text-ink disabled:opacity-50"
                      >
                        {isEditing ? <X className="h-3.5 w-3.5" /> : <Pencil className="h-3.5 w-3.5" />}
                        {isEditing ? "Tutup" : "Koreksi"}
                      </button>
                    </div>
                  )}
                </div>

                {/* Correction panel — set any status directly, points follow */}
                {isEditing && !v.is_void && (
                  <div className="mt-4 rounded-xl border border-brand-100 bg-brand-50/40 p-3">
                    <p className="text-xs font-black uppercase tracking-wider text-muted">Koreksi status</p>
                    <div className="mt-2 flex flex-wrap gap-2">
                      {STATUS_ORDER.map((s) => {
                        const active = s === v.status_hadir;
                        const delta = STATUS_CUMULATIVE[s] - STATUS_CUMULATIVE[v.status_hadir];
                        return (
                          <button
                            key={s}
                            onClick={() => changeStatus(v, s)}
                            disabled={active || busy}
                            className={`rounded-xl px-3 py-2 text-xs font-black transition ${
                              active
                                ? "bg-brand-600 text-white"
                                : "border border-brand-200 bg-white text-ink hover:border-brand-400 disabled:opacity-50"
                            }`}
                          >
                            {STATUS_LABEL[s]}
                            {!active && <span className={`ml-1.5 ${delta < 0 ? "text-red-600" : "text-green-600"}`}>{formatDelta(delta)}</span>}
                          </button>
                        );
                      })}
                    </div>

                    {v.is_converted && (
                      <>
                        <p className="mt-4 text-xs font-black uppercase tracking-wider text-muted">Koreksi konversi</p>
                        <button
                          onClick={() => changeConversion(v, false)}
                          disabled={busy}
                          className="mt-2 rounded-xl border border-red-200 bg-white px-3 py-2 text-xs font-black text-red-700 hover:bg-red-50 disabled:opacity-50"
                        >
                          Batalkan konversi (−{CONVERSION_POINTS} pts)
                        </button>
                      </>
                    )}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}
    </>
  );
}
