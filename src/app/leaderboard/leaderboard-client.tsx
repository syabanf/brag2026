"use client";

import { BarChart3, Check, Copy, Share2, Trophy, X } from "lucide-react";
import { useState } from "react";
import { createPortal } from "react-dom";
import { formatCurrencyCompact, formatPoints } from "@/lib/utils";
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
          <div className="space-y-2.5">
            {sortedTeams.map((team, index) => {
              // Top three get a restrained metal accent — a tinted rank badge
              // and hairline edge — instead of a full gradient card, so the
              // numbers stay the loudest thing on the row.
              const accent =
                index === 0 ? { edge: "border-amber-300", rank: "bg-amber-400 text-amber-950", score: "text-amber-700" } :
                index === 1 ? { edge: "border-slate-300", rank: "bg-slate-300 text-slate-800", score: "text-slate-700" } :
                index === 2 ? { edge: "border-orange-300", rank: "bg-orange-300 text-orange-950", score: "text-orange-800" } :
                { edge: "", rank: "bg-brand-50 text-brand-700", score: "text-brand-600" };
              return (
              <div
                className={`card p-3 ${accent.edge ? `border-l-[3px] ${accent.edge}` : ""}`}
                key={team.team_id}
              >
                <div className="flex items-center justify-between gap-2.5">
                  <div className="flex min-w-0 items-center gap-2.5">
                    <span className={`num grid h-9 w-9 shrink-0 place-items-center rounded-xl text-sm font-black ${accent.rank}`}>
                      {index + 1}
                    </span>
                    <p className="truncate text-base font-black text-ink">{team.nama_tim}</p>
                  </div>
                  <div className="shrink-0 text-right">
                    <p className={`num text-lg font-black leading-none ${accent.score}`}>
                      {formatTeamTabScore(team, tab)}
                    </p>
                    {formatTeamTabSubScore(team, tab) && (
                      <p className="num mt-0.5 text-[0.68rem] font-bold text-muted">
                        {formatTeamTabSubScore(team, tab)}
                      </p>
                    )}
                  </div>
                </div>

                <div className="mt-2.5 grid grid-cols-2 gap-2 text-xs font-bold">
                  <button
                    type="button"
                    onClick={() => setHistory({ team, kind: "tyfcb" })}
                    aria-label={`Lihat riwayat TYFCB ${team.nama_tim}`}
                    title={`Rp ${Number(team.nilai_tyfcb).toLocaleString("id-ID")}`}
                    className="flex min-h-11 flex-col justify-center rounded-xl bg-brand-50/70 px-2.5 py-1.5 text-left transition hover:bg-brand-50 active:scale-[0.98]"
                  >
                    <span className="num text-ink">{formatCurrencyCompact(Number(team.nilai_tyfcb))}</span>
                    <span className="num text-[0.65rem] font-semibold text-muted">{team.count_tyfcb}× transaksi TYFCB</span>
                  </button>
                  <button
                    type="button"
                    onClick={() => setHistory({ team, kind: "visitor" })}
                    aria-label={`Lihat riwayat Visitor ${team.nama_tim}`}
                    className="flex min-h-11 flex-col justify-center rounded-xl bg-brand-50/70 px-2.5 py-1.5 text-left transition hover:bg-brand-50 active:scale-[0.98]"
                  >
                    <span className="num text-ink">{team.count_visitor} visitor</span>
                    <span className="text-[0.65rem] font-semibold text-muted">Tamu diundang</span>
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
  const [open, setOpen] = useState(false);
  const [copied, setCopied] = useState(false);
  // Resolved on click rather than in an effect: the modal only ever exists
  // after a user gesture, so window is guaranteed to be available by then.
  const [url, setUrl] = useState("");

  function openDialog() {
    setUrl(`${window.location.origin}/public/leaderboard`);
    setOpen(true);
  }

  async function copyLink() {
    await navigator.clipboard.writeText(url);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  const modal = open ? createPortal(
    <div
      className="fixed inset-0 z-[9999] flex items-center justify-center bg-ink/40 px-4 backdrop-blur-sm"
      onClick={(e) => { if (e.target === e.currentTarget) setOpen(false); }}
    >
      <div className="w-full max-w-sm rounded-2xl bg-white p-5 shadow-2xl">
        <div className="mb-3 flex items-center justify-between gap-3">
          <h3 className="text-base font-black text-ink">Bagikan Leaderboard</h3>
          <button
            type="button"
            aria-label="Tutup"
            onClick={() => setOpen(false)}
            className="grid h-9 w-9 shrink-0 place-items-center rounded-full text-muted transition hover:bg-brand-50 hover:text-ink"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <p className="mb-3 text-sm leading-relaxed text-muted">
          Siapa pun bisa melihat posisi team tanpa perlu login.
        </p>

        <div className="flex items-center gap-2 rounded-xl border border-brand-100 bg-brand-50 px-3 py-2.5">
          <p className="min-w-0 flex-1 truncate text-sm font-semibold text-ink">{url}</p>
        </div>

        <button
          type="button"
          onClick={copyLink}
          className={`mt-3 flex min-h-12 w-full items-center justify-center gap-2 rounded-xl text-sm font-black transition active:scale-[0.98] ${
            copied
              ? "bg-emerald-600 text-white"
              : "bg-brand-600 text-white hover:bg-brand-700"
          }`}
        >
          {copied ? (
            <><Check className="h-4 w-4" />Link tersalin</>
          ) : (
            <><Copy className="h-4 w-4" />Copy link</>
          )}
        </button>
      </div>
    </div>,
    document.body
  ) : null;

  return (
    <>
      <button
        type="button"
        onClick={openDialog}
        className="flex shrink-0 items-center gap-1.5 rounded-full border border-brand-100 bg-white px-3 py-2 text-xs font-bold text-brand-600 transition hover:bg-brand-50 active:scale-95"
      >
        <Share2 className="h-3.5 w-3.5" />
        Bagikan
      </button>
      {modal}
    </>
  );
}
