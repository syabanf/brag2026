"use client";

import { BarChart3, Check, Copy, Share2, Trophy, X } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { formatPoints } from "@/lib/utils";
import { TeamHistoryDialog, type HistoryKind } from "@/components/team-history-dialog";

type Tab = "overall" | "tyfcb" | "visitor";

export type TeamRow = {
  team_id: string;
  nama_tim: string;
  score_overall: number;
  nilai_tyfcb: number;    // collective sum of verified tyfcb_entries.nilai (Rp)
  count_tyfcb: number;    // number of verified tyfcb_entries behind nilai_tyfcb
  count_visitor: number;  // count of visitors invited by team members
};

export type MemberRow = {
  id: string;
  full_name: string;
  nama_tim: string | null;
  klasifikasi_nama: string | null;
  color_status: string;
  score_overall: number;
  score_tyfcb: number;
  score_visitor: number;
};

const TABS: { key: Tab; label: string }[] = [
  { key: "overall", label: "Overall" },
  { key: "tyfcb",   label: "TYFCB" },
  { key: "visitor", label: "Visitor" },
];

function teamSortKey(t: TeamRow, tab: Tab): number {
  if (tab === "tyfcb")   return Number(t.nilai_tyfcb);
  if (tab === "visitor") return t.count_visitor;
  return t.score_overall;
}

function memberScore(m: MemberRow, tab: Tab) {
  if (tab === "tyfcb")   return m.score_tyfcb;
  if (tab === "visitor") return m.score_visitor;
  return m.score_overall;
}

function formatTeamTabScore(team: TeamRow, tab: Tab): string {
  if (tab === "tyfcb")   return `Rp ${Number(team.nilai_tyfcb).toLocaleString("id-ID")}`;
  if (tab === "visitor") return `${team.count_visitor} visitor`;
  return `${formatPoints(team.score_overall)} pts`;
}

function formatTeamTabSubScore(team: TeamRow, tab: Tab): string | null {
  if (tab === "tyfcb") return `${team.count_tyfcb}× TYFCB`;
  return null;
}

export function LeaderboardClient({
  teams,
  members,
}: {
  teams: TeamRow[];
  members: MemberRow[];
}) {
  const [tab, setTab] = useState<Tab>("overall");
  const [history, setHistory] = useState<{ team: TeamRow; kind: HistoryKind } | null>(null);

  const sortedTeams   = [...teams].sort((a, b) => teamSortKey(b, tab) - teamSortKey(a, tab));
  const sortedMembers = [...members].sort((a, b) => memberScore(b, tab) - memberScore(a, tab));

  return (
    <>
      {/* Tab bar */}
      <div className="flex rounded-full border border-brand-100 bg-white p-1 text-sm font-bold text-muted">
        {TABS.map(({ key, label }) => (
          <button
            key={key}
            onClick={() => setTab(key)}
            className={`rounded-full px-4 py-2 transition ${
              tab === key ? "bg-brand-600 text-white" : "hover:text-ink"
            }`}
          >
            {label}
          </button>
        ))}
      </div>

      <div className="mt-6 grid gap-6 lg:grid-cols-2">
        {/* Team leaderboard */}
        <section className="glass-panel rounded-2xl p-4">
          <div className="mb-4 flex items-center justify-between gap-3">
            <div className="flex items-center gap-3">
              <Trophy className="h-6 w-6 text-brand-600" />
              <h2 className="text-xl font-black">
                Team {tab === "overall" ? "Overall" : tab.toUpperCase()}
              </h2>
            </div>
            {tab === "overall" && <ShareTeamButton />}
          </div>
          <div className="space-y-3">
            {sortedTeams.map((team, index) => {
              const cardCls =
                index === 0 ? "bg-gradient-to-br from-yellow-300 via-amber-400 to-orange-400 border-yellow-200 shadow-lift" :
                index === 1 ? "bg-gradient-to-br from-slate-300 via-gray-200 to-slate-400 border-slate-200 shadow-lift" :
                index === 2 ? "bg-gradient-to-br from-amber-700 via-amber-600 to-orange-700 border-amber-500 shadow-lift" :
                "border border-brand-100 bg-white";
              const nameCls  = index === 0 ? "text-yellow-900" : index === 1 ? "text-slate-800" : index === 2 ? "text-white" : "text-ink";
              const scoreCls = index === 0 ? "text-yellow-900" : index === 1 ? "text-slate-800" : index === 2 ? "text-white" : "text-brand-600";
              const chipCls  = index === 0 ? "bg-white/30 text-yellow-900" : index === 1 ? "bg-white/30 text-slate-800" : index === 2 ? "bg-white/20 text-white/90" : "bg-brand-50 text-muted";
              const rankCls  = index === 0 ? "bg-white/40 text-yellow-900" : index === 1 ? "bg-white/40 text-slate-800" : index === 2 ? "bg-white/25 text-white" : "bg-brand-600 text-white";
              return (
              <div className={`rounded-2xl p-4 ${cardCls}`} key={team.team_id}>
                <div className="flex items-center justify-between gap-3">
                  <div className="flex min-w-0 items-center gap-3">
                    <span className={`grid h-12 w-12 shrink-0 place-items-center rounded-full text-lg font-black ${rankCls}`}>
                      {index + 1}
                    </span>
                    <p className={`truncate text-lg font-black ${nameCls}`}>{team.nama_tim}</p>
                  </div>
                  <div className="shrink-0 text-right">
                    <p className={`text-xl font-black ${scoreCls}`}>
                      {formatTeamTabScore(team, tab)}
                    </p>
                    {formatTeamTabSubScore(team, tab) && (
                      <p className={`text-xs font-bold opacity-70 ${scoreCls}`}>
                        {formatTeamTabSubScore(team, tab)}
                      </p>
                    )}
                  </div>
                </div>
                <div className="mt-4 grid grid-cols-2 gap-2 text-center text-xs font-bold">
                  <button
                    type="button"
                    onClick={() => setHistory({ team, kind: "tyfcb" })}
                    aria-label={`Lihat riwayat TYFCB ${team.nama_tim}`}
                    className={`flex flex-col justify-center rounded-xl p-2 transition hover:brightness-105 active:scale-[0.97] ${chipCls}`}
                  >
                    <span>TYFCB Rp {Number(team.nilai_tyfcb).toLocaleString("id-ID")}</span>
                    <span className="opacity-70">{team.count_tyfcb}× transaksi</span>
                  </button>
                  <button
                    type="button"
                    onClick={() => setHistory({ team, kind: "visitor" })}
                    aria-label={`Lihat riwayat Visitor ${team.nama_tim}`}
                    className={`flex flex-col justify-center rounded-xl p-2 transition hover:brightness-105 active:scale-[0.97] ${chipCls}`}
                  >
                    Visitor {team.count_visitor}
                  </button>
                </div>
              </div>
              );
            })}
          </div>
        </section>

        {/* Individual leaderboard */}
        <section className="glass-panel rounded-2xl p-4">
          <div className="mb-4 flex items-center gap-3">
            <BarChart3 className="h-6 w-6 text-brand-600" />
            <h2 className="text-xl font-black">
              Individu {tab === "overall" ? "Overall" : tab.toUpperCase()}
            </h2>
          </div>
          <div className="space-y-2">
            {sortedMembers.map((m, index) => (
              <div
                className="grid grid-cols-[2rem_2.75rem_1fr_auto] items-center gap-3 rounded-xl bg-white px-3 py-3"
                key={m.id}
              >
                <span className="text-center text-sm font-black text-muted">{index + 1}</span>
                <span className="grid h-11 w-11 place-items-center rounded-full bg-brand-50 text-xs font-bold text-brand-700">
                  {m.full_name.split(" ").map((n) => n[0]).join("").slice(0, 2).toUpperCase()}
                </span>
                <div className="min-w-0">
                  <p className="truncate font-bold text-ink">{m.full_name}</p>
                  <p className="truncate text-xs text-muted">
                    {m.nama_tim ?? "—"}{m.klasifikasi_nama ? ` · ${m.klasifikasi_nama}` : ""}
                  </p>
                </div>
                <div className="text-right">
                  <p className="font-black text-brand-600">{formatPoints(memberScore(m, tab))} pts</p>
                  <p className="text-xs text-muted capitalize">{m.color_status}</p>
                </div>
              </div>
            ))}
          </div>
        </section>
      </div>

      {history && (
        <TeamHistoryDialog
          teamId={history.team.team_id}
          teamName={history.team.nama_tim}
          kind={history.kind}
          onClose={() => setHistory(null)}
        />
      )}
    </>
  );
}

function ShareTeamButton() {
  const [open, setOpen]     = useState(false);
  const [copied, setCopied] = useState(false);
  const [mounted, setMounted] = useState(false);
  const urlRef = useRef("");

  useEffect(() => {
    setMounted(true);
    urlRef.current = `${window.location.origin}/public/leaderboard`;
  }, []);

  async function copyLink() {
    await navigator.clipboard.writeText(urlRef.current);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  const modal = mounted && open ? createPortal(
    <div
      className="fixed inset-0 z-[9999] flex items-center justify-center bg-black/40 px-4 backdrop-blur-sm"
      onClick={(e) => { if (e.target === e.currentTarget) setOpen(false); }}
    >
      <div className="w-full max-w-sm rounded-2xl bg-white p-6 shadow-2xl">
        <div className="mb-4 flex items-center justify-between">
          <h3 className="text-lg font-black text-ink">Bagikan Leaderboard</h3>
          <button
            onClick={() => setOpen(false)}
            className="grid h-8 w-8 place-items-center rounded-full text-muted hover:bg-brand-50 hover:text-ink"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <p className="mb-3 text-sm text-muted">
          Siapapun bisa melihat posisi team tanpa perlu login.
        </p>

        <div className="flex items-center gap-2 rounded-xl border border-brand-100 bg-brand-50 px-3 py-2.5">
          <p className="min-w-0 flex-1 truncate text-sm font-semibold text-ink">{urlRef.current}</p>
        </div>

        <button
          onClick={copyLink}
          className={`mt-3 flex w-full items-center justify-center gap-2 rounded-xl py-3 text-sm font-black transition active:scale-[0.98] ${
            copied
              ? "bg-green-600 text-white"
              : "bg-brand-600 text-white hover:bg-brand-700"
          }`}
        >
          {copied ? (
            <><Check className="h-4 w-4" />Link Tersalin!</>
          ) : (
            <><Copy className="h-4 w-4" />Copy Link</>
          )}
        </button>
      </div>
    </div>,
    document.body
  ) : null;

  return (
    <>
      <button
        onClick={() => setOpen(true)}
        className="flex items-center gap-1.5 rounded-full border border-brand-200 bg-white px-3 py-1.5 text-xs font-bold text-brand-600 transition hover:bg-brand-50 active:scale-95"
      >
        <Share2 className="h-3.5 w-3.5" />
        Bagikan
      </button>
      {modal}
    </>
  );
}
