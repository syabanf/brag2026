"use client";

import { CheckCircle2, Clock, XCircle, Zap } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";

export type TyfcbRow = {
  id: string;
  giver_name: string;
  receiver_name: string;
  nilai: string;
  tanggal: string;
  status: "pending" | "verified" | "rejected";
  computed_score: number | null;
  /** Numeric string from the DB; > 1 means a booster multiplied this entry's score. */
  event_multiplier_applied: string | null;
  /** Null when no active booster still covers the entry — show the multiplier alone. */
  booster_judul: string | null;
};

type Filter = "all" | "pending" | "verified" | "rejected";

/** "2.00" -> 2, so the badge reads "2x" and not "2.00x". Null when no booster applied. */
function boosterMultiplier(entry: TyfcbRow): number | null {
  const m = Number(entry.event_multiplier_applied);
  return Number.isFinite(m) && m > 1 ? m : null;
}

const STATUS_STYLE = {
  pending:  "bg-yellow-50 text-yellow-700",
  verified: "bg-green-50 text-green-700",
  rejected: "bg-red-50 text-red-700",
};
const STATUS_LABEL = {
  pending:  "Pending",
  verified: "Approved",
  rejected: "Rejected",
};

export function TyfcbAdminClient({ initial }: { initial: TyfcbRow[] }) {
  const router = useRouter();
  const [entries, setEntries] = useState(initial);
  const [filter, setFilter] = useState<Filter>("pending");
  const [loading, setLoading] = useState<string | null>(null);
  const [error, setError] = useState("");

  async function updateStatus(id: string, status: "pending" | "verified" | "rejected") {
    setLoading(id + status);
    setError("");
    const res = await fetch(`/api/admin/tyfcb/${id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ status }),
    });
    const d = await res.json();
    setLoading(null);
    if (!res.ok) { setError(d.error ?? "Gagal."); return; }
    setEntries((prev) => prev.map((e) => e.id === id ? { ...e, status } : e));
    router.refresh();
  }

  const counts = {
    all: entries.length,
    pending: entries.filter((e) => e.status === "pending").length,
    verified: entries.filter((e) => e.status === "verified").length,
    rejected: entries.filter((e) => e.status === "rejected").length,
  };

  const visible = filter === "all" ? entries : entries.filter((e) => e.status === filter);

  return (
    <>
      {/* Filter tabs */}
      <div className="mb-4 flex gap-2 overflow-x-auto">
        {(["pending", "all", "verified", "rejected"] as Filter[]).map((f) => (
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
          Tidak ada submission.
        </div>
      ) : (
        <div className="space-y-3">
          {visible.map((e) => (
            <div key={e.id} className="rounded-2xl border border-brand-100 bg-white p-4">
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div className="min-w-0 flex-1">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="text-xs text-muted">Pembeli:</span>
                    <p className="font-black text-ink">{e.giver_name}</p>
                    <span className="text-xs text-muted">· Penjual:</span>
                    <p className="font-bold text-ink">{e.receiver_name}</p>
                    <span className={`rounded-full px-2.5 py-0.5 text-xs font-black ${STATUS_STYLE[e.status]}`}>
                      {STATUS_LABEL[e.status]}
                    </span>
                  </div>
                  <p className="mt-1 text-sm text-muted">
                    Rp {Number(e.nilai).toLocaleString("id-ID")} · {e.tanggal}
                    {e.computed_score != null && (
                      <span className="ml-2 font-bold text-brand-700">+{e.computed_score} pts</span>
                    )}
                  </p>
                  {boosterMultiplier(e) != null && (
                    <span className="mt-2 inline-flex max-w-full items-center gap-1.5 rounded-full bg-orange-50 px-2.5 py-1 text-xs font-black text-orange-700 ring-1 ring-orange-200">
                      <Zap className="h-3.5 w-3.5 shrink-0" />
                      <span className="shrink-0">{boosterMultiplier(e)}x booster</span>
                      {e.booster_judul && (
                        <span className="truncate font-bold text-orange-600">· {e.booster_judul}</span>
                      )}
                    </span>
                  )}
                </div>

                {/* Action buttons */}
                <div className="flex shrink-0 gap-2">
                  {e.status !== "verified" && (
                    <button
                      onClick={() => updateStatus(e.id, "verified")}
                      disabled={loading === e.id + "verified"}
                      className="flex items-center gap-1.5 rounded-xl bg-green-600 px-3 py-2 text-xs font-black text-white hover:bg-green-700 disabled:opacity-50"
                    >
                      <CheckCircle2 className="h-3.5 w-3.5" />
                      Approve
                    </button>
                  )}
                  {e.status !== "pending" && (
                    <button
                      onClick={() => updateStatus(e.id, "pending")}
                      disabled={loading === e.id + "pending"}
                      className="flex items-center gap-1.5 rounded-xl border border-yellow-300 bg-yellow-50 px-3 py-2 text-xs font-black text-yellow-700 hover:bg-yellow-100 disabled:opacity-50"
                    >
                      <Clock className="h-3.5 w-3.5" />
                      Pending
                    </button>
                  )}
                  {e.status !== "rejected" && (
                    <button
                      onClick={() => updateStatus(e.id, "rejected")}
                      disabled={loading === e.id + "rejected"}
                      className="flex items-center gap-1.5 rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-xs font-black text-red-600 hover:bg-red-100 disabled:opacity-50"
                    >
                      <XCircle className="h-3.5 w-3.5" />
                      Reject
                    </button>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </>
  );
}
