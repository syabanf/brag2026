"use client";

import { Gift, Loader2, UserPlus, X } from "lucide-react";
import { useEffect, useState } from "react";
import { createPortal } from "react-dom";
import type { TeamHistoryResponse } from "@/lib/domain/types";

export type HistoryKind = "tyfcb" | "visitor";

const VISITOR_STATUS_STYLE: Record<string, string> = {
  terdaftar:   "bg-slate-50 text-slate-600",
  hadir:       "bg-blue-50 text-blue-700",
  hadir_penuh: "bg-green-50 text-green-700",
};

const VISITOR_STATUS_LABEL: Record<string, string> = {
  terdaftar:   "Terdaftar",
  hadir:       "Hadir",
  hadir_penuh: "Hadir Penuh",
};

// `scope` picks the API surface: "member" needs a session, "public" does not.
export type HistoryScope = "member" | "public";

const HISTORY_ENDPOINT: Record<HistoryScope, (teamId: string) => string> = {
  member: (id) => `/api/leaderboard/teams/${id}/history`,
  public: (id) => `/api/public/leaderboard/teams/${id}/history`,
};

export function TeamHistoryDialog({
  teamId,
  teamName,
  kind,
  scope = "member",
  onClose,
}: {
  teamId: string;
  teamName: string;
  kind: HistoryKind;
  scope?: HistoryScope;
  onClose: () => void;
}) {
  const [data, setData]   = useState<TeamHistoryResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const controller = new AbortController();

    async function load() {
      setData(null);
      setError(null);
      try {
        const res = await fetch(HISTORY_ENDPOINT[scope](teamId), {
          signal: controller.signal,
        });
        if (!res.ok) {
          const body = await res.json().catch(() => null);
          throw new Error(body?.error ?? "Gagal memuat riwayat.");
        }
        setData(await res.json());
      } catch (err: unknown) {
        if (err instanceof Error && err.name === "AbortError") return;
        setError(err instanceof Error ? err.message : "Gagal memuat riwayat.");
      }
    }

    load();
    return () => controller.abort();
  }, [teamId, scope]);

  useEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape") onClose();
    }
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [onClose]);

  const isTyfcb    = kind === "tyfcb";
  const Icon       = isTyfcb ? Gift : UserPlus;
  const totalPoints = (data?.tyfcb ?? []).reduce((sum, e) => sum + (e.computed_score ?? 0), 0);

  return createPortal(
    <div
      className="fixed inset-0 z-[9999] flex items-end justify-center bg-black/40 backdrop-blur-sm sm:items-center sm:px-4"
      onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}
      role="dialog"
      aria-modal="true"
      aria-label={`Riwayat ${isTyfcb ? "TYFCB" : "Visitor"} ${teamName}`}
    >
      <div className="flex max-h-[85vh] w-full max-w-lg flex-col rounded-t-2xl bg-white shadow-2xl sm:rounded-2xl">
        {/* Header */}
        <div className="flex items-start justify-between gap-3 border-b border-brand-100 p-5">
          <div className="flex min-w-0 items-center gap-3">
            <span className="grid h-10 w-10 shrink-0 place-items-center rounded-full bg-brand-50">
              <Icon className="h-5 w-5 text-brand-600" />
            </span>
            <div className="min-w-0">
              <h3 className="truncate text-lg font-black text-ink">
                Riwayat {isTyfcb ? "TYFCB" : "Visitor"}
              </h3>
              <p className="truncate text-sm text-muted">{teamName}</p>
            </div>
          </div>
          <button
            onClick={onClose}
            aria-label="Tutup"
            className="grid h-8 w-8 shrink-0 place-items-center rounded-full text-muted hover:bg-brand-50 hover:text-ink"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        {/* Summary */}
        {data && (
          <div className="flex items-baseline gap-2 border-b border-brand-50 bg-brand-50/40 px-5 py-3">
            <span className="text-xl font-black text-brand-600">
              {isTyfcb
                ? `${totalPoints.toLocaleString("id-ID")} pts`
                : `${data.visitors.length} visitor`}
            </span>
            {isTyfcb && (
              <span className="text-sm font-bold text-muted">
                dari {data.tyfcb.length}× transaksi
              </span>
            )}
          </div>
        )}

        {/* Body */}
        <div className="min-h-0 flex-1 overflow-y-auto p-4">
          {error ? (
            <p className="py-10 text-center text-sm font-semibold text-red-600">{error}</p>
          ) : !data ? (
            <div className="flex items-center justify-center gap-2 py-10 text-muted">
              <Loader2 className="h-4 w-4 animate-spin" />
              <span className="text-sm font-semibold">Memuat riwayat…</span>
            </div>
          ) : (isTyfcb ? data.tyfcb.length : data.visitors.length) === 0 ? (
            <p className="py-10 text-center text-sm text-muted">
              Belum ada {isTyfcb ? "TYFCB terverifikasi" : "visitor"} untuk team ini.
            </p>
          ) : isTyfcb ? (
            <ul className="space-y-2">
              {data.tyfcb.map((e) => (
                <li
                  key={e.id}
                  className="flex items-start justify-between gap-3 rounded-xl border border-brand-100 p-3"
                >
                  <div className="min-w-0">
                    <p className="truncate font-bold text-ink">{e.buyer_name}</p>
                    <p className="truncate text-xs text-muted">
                      Penjual: {e.seller_name} · {e.tanggal}
                    </p>
                  </div>
                  <div className="shrink-0 text-right">
                    <p className="font-black text-brand-600">
                      {e.computed_score != null ? `+${e.computed_score} pts` : "—"}
                    </p>
                  </div>
                </li>
              ))}
            </ul>
          ) : (
            <ul className="space-y-2">
              {data.visitors.map((v) => (
                <li
                  key={v.id}
                  className="flex items-start justify-between gap-3 rounded-xl border border-brand-100 p-3"
                >
                  <div className="min-w-0">
                    <p className="truncate font-bold text-ink">{v.nama}</p>
                    <p className="truncate text-xs text-muted">
                      Diundang {v.inviter_name} · {v.tanggal_undang}
                    </p>
                  </div>
                  <div className="flex shrink-0 flex-col items-end gap-1">
                    <span
                      className={`rounded-full px-2.5 py-0.5 text-xs font-bold ${
                        VISITOR_STATUS_STYLE[v.status_hadir] ?? "bg-slate-50 text-slate-600"
                      }`}
                    >
                      {VISITOR_STATUS_LABEL[v.status_hadir] ?? v.status_hadir}
                    </span>
                    {v.is_converted && (
                      <span className="rounded-full bg-green-50 px-2.5 py-0.5 text-xs font-bold text-green-700">
                        Converted
                      </span>
                    )}
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>
    </div>,
    document.body
  );
}
