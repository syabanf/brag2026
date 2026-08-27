"use client";

import { CheckCircle2, Clock, XCircle } from "lucide-react";
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
};

type Filter = "all" | "pending" | "verified" | "rejected";

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
      <div className="no-scrollbar mb-4 flex gap-2 overflow-x-auto">
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
            <div key={e.id} className="card p-3.5">
              <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                <div className="min-w-0 flex-1">
                  <div className="flex flex-wrap items-center gap-x-2 gap-y-1">
                    <span className={`rounded-full px-2.5 py-0.5 text-[0.68rem] font-black ${STATUS_STYLE[e.status]}`}>
                      {STATUS_LABEL[e.status]}
                    </span>
                    <p className="num text-sm font-black text-ink">
                      Rp {Number(e.nilai).toLocaleString("id-ID")}
                    </p>
                    <span className="text-xs text-muted">· {e.tanggal}</span>
                    {e.computed_score != null && (
                      <span className="num text-xs font-bold text-brand-700">+{e.computed_score} pts</span>
                    )}
                  </div>

                  <dl className="mt-2 space-y-0.5 text-sm">
                    <div className="flex gap-2">
                      <dt className="w-14 shrink-0 text-xs leading-5 text-muted">Pembeli</dt>
                      <dd className="min-w-0 truncate font-bold text-ink">{e.giver_name}</dd>
                    </div>
                    <div className="flex gap-2">
                      <dt className="w-14 shrink-0 text-xs leading-5 text-muted">Penjual</dt>
                      <dd className="min-w-0 truncate font-bold text-ink">{e.receiver_name}</dd>
                    </div>
                  </dl>
                </div>

                <div className="grid grid-cols-2 gap-2 lg:flex lg:shrink-0">
                  {e.status !== "verified" && (
                    <button
                      onClick={() => updateStatus(e.id, "verified")}
                      disabled={loading === e.id + "verified"}
                      className="flex min-h-11 items-center justify-center gap-1.5 rounded-xl bg-emerald-600 px-3 text-xs font-black text-white transition hover:bg-emerald-700 disabled:opacity-50"
                    >
                      <CheckCircle2 className="h-3.5 w-3.5" />
                      Approve
                    </button>
                  )}
                  {e.status !== "pending" && (
                    <button
                      onClick={() => updateStatus(e.id, "pending")}
                      disabled={loading === e.id + "pending"}
                      className="flex min-h-11 items-center justify-center gap-1.5 rounded-xl border border-amber-300 bg-amber-50 px-3 text-xs font-black text-amber-700 transition hover:bg-amber-100 disabled:opacity-50"
                    >
                      <Clock className="h-3.5 w-3.5" />
                      Pending
                    </button>
                  )}
                  {e.status !== "rejected" && (
                    <button
                      onClick={() => updateStatus(e.id, "rejected")}
                      disabled={loading === e.id + "rejected"}
                      className="flex min-h-11 items-center justify-center gap-1.5 rounded-xl border border-red-200 bg-red-50 px-3 text-xs font-black text-red-600 transition hover:bg-red-100 disabled:opacity-50"
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
