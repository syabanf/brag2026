import { useState } from "react";
import { BarChart3, Trophy, X } from "lucide-react";
import { api } from "../lib/api";
import { useApi } from "../lib/use-api";
import { formatCurrency, formatCurrencyCompact, formatDate, formatPoints } from "../lib/format";
import { EmptyState, ErrorNote, PageHeader, Spinner, Tabs } from "../components/ui";
import type { LedgerEntry, MemberScore, TeamScore } from "../lib/types";

type Tab = "overall" | "tyfcb" | "visitor";

const TABS: { key: Tab; label: string }[] = [
  { key: "overall", label: "Overall" },
  { key: "tyfcb", label: "TYFCB" },
  { key: "visitor", label: "Visitor" },
];

export function LeaderboardPage({ isPublic = false }: { isPublic?: boolean }) {
  const [tab, setTab] = useState<Tab>("overall");
  const [history, setHistory] = useState<{ team: TeamScore; kind: "tyfcb" | "visitor" } | null>(null);

  // The member list is truncated server-side, so switching tabs refetches:
  // re-sorting a top-50-by-overall would answer "who leads on visitors" with
  // the wrong fifty people. Team scores are complete and sort on the client.
  const { data, error, loading, reload } = useApi(
    () => (isPublic ? api.leaderboard.public(tab) : api.leaderboard.get(tab)),
    [isPublic, tab],
  );

  if (loading) return <Spinner />;
  if (error) return <ErrorNote message={error} onRetry={reload} />;
  if (!data) return null;

  const teams = [...data.teams].sort((a, b) => scoreFor(b, tab) - scoreFor(a, tab));

  return (
    <div className="space-y-5">
      <PageHeader
        title="Leaderboard"
        description="Pantau posisi tim di tiap kategori — poin keseluruhan, nilai TYFCB kolektif, dan jumlah visitor yang berhasil diundang."
      />

      <Tabs tabs={TABS} active={tab} onChange={setTab} />

      <section className="card p-4">
        <div className="mb-3 flex items-center gap-2.5">
          <Trophy className="h-5 w-5 text-brand-600" />
          <h2 className="text-base font-black text-ink">
            Team {TABS.find((t) => t.key === tab)?.label}
          </h2>
        </div>

        <div className="space-y-2.5">
          {teams.map((team, index) => {
            // Top three get a restrained metal accent rather than a full
            // gradient card, so the numbers stay the loudest thing on the row.
            const accent =
              index === 0
                ? { edge: "border-l-[3px] border-amber-300", rank: "bg-amber-400 text-amber-950", score: "text-amber-700" }
                : index === 1
                  ? { edge: "border-l-[3px] border-slate-300", rank: "bg-slate-300 text-slate-800", score: "text-slate-700" }
                  : index === 2
                    ? { edge: "border-l-[3px] border-orange-300", rank: "bg-orange-300 text-orange-950", score: "text-orange-800" }
                    : { edge: "", rank: "bg-brand-50 text-brand-700", score: "text-brand-600" };

            return (
              <div key={team.team_id} className={`card p-3 ${accent.edge}`}>
                <div className="flex items-center justify-between gap-2.5">
                  <div className="flex min-w-0 items-center gap-2.5">
                    <span className={`num grid h-9 w-9 shrink-0 place-items-center rounded-xl text-sm font-black ${accent.rank}`}>
                      {index + 1}
                    </span>
                    <p className="truncate text-base font-black text-ink">{team.nama_tim}</p>
                  </div>
                  <p className={`num shrink-0 text-lg font-black leading-none ${accent.score}`}>
                    {labelFor(team, tab)}
                  </p>
                </div>

                <div className="mt-2.5 grid grid-cols-2 gap-2 text-xs font-bold">
                  <button
                    type="button"
                    onClick={() => setHistory({ team, kind: "tyfcb" })}
                    title={formatCurrency(team.nilai_tyfcb)}
                    className="flex min-h-11 flex-col justify-center rounded-xl bg-brand-50/70 px-2.5 py-1.5 text-left transition hover:bg-brand-50 active:scale-[0.98]"
                  >
                    <span className="num text-ink">{formatCurrencyCompact(team.nilai_tyfcb)}</span>
                    <span className="num text-[0.65rem] font-semibold text-muted">
                      {team.count_tyfcb}× transaksi TYFCB
                    </span>
                  </button>
                  <button
                    type="button"
                    onClick={() => setHistory({ team, kind: "visitor" })}
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

      {!isPublic && data.members.length > 0 && (
        <section className="card p-4">
          <div className="mb-3 flex items-center gap-2.5">
            <BarChart3 className="h-5 w-5 text-brand-600" />
            <h2 className="text-base font-black text-ink">
              Individu {TABS.find((t) => t.key === tab)?.label}
            </h2>
          </div>

          <ol className="space-y-1.5">
            {data.members.map((member, index) => (
              <li
                key={member.member_id}
                className="grid grid-cols-[1.75rem_1fr_auto] items-center gap-2.5 rounded-xl px-2 py-2"
              >
                <span className="num text-center text-sm font-black text-muted">{index + 1}</span>
                <div className="min-w-0">
                  <p className="truncate text-sm font-bold text-ink">{member.full_name}</p>
                  <p className="truncate text-xs text-muted">{member.nama_tim ?? "—"}</p>
                </div>
                <span className="num text-sm font-black text-brand-600">
                  {formatPoints(memberScoreFor(member, tab))}
                </span>
              </li>
            ))}
          </ol>
        </section>
      )}

      {history && (
        <HistoryDialog
          team={history.team}
          kind={history.kind}
          isPublic={isPublic}
          onClose={() => setHistory(null)}
        />
      )}
    </div>
  );
}

function scoreFor(team: TeamScore, tab: Tab) {
  if (tab === "tyfcb") return team.score_tyfcb;
  if (tab === "visitor") return team.score_visitor;
  return team.score_overall;
}

function memberScoreFor(member: MemberScore, tab: Tab) {
  if (tab === "tyfcb") return member.score_tyfcb;
  if (tab === "visitor") return member.score_visitor;
  return member.score_overall;
}

function labelFor(team: TeamScore, tab: Tab) {
  return `${formatPoints(scoreFor(team, tab))} pts`;
}

function HistoryDialog({
  team,
  kind,
  isPublic,
  onClose,
}: {
  team: TeamScore;
  kind: "tyfcb" | "visitor";
  isPublic: boolean;
  onClose: () => void;
}) {
  const { data, error, loading } = useApi<LedgerEntry[]>(
    () =>
      isPublic
        ? api.leaderboard.publicTeamHistory(team.team_id, kind)
        : api.leaderboard.teamHistory(team.team_id, kind),
    [team.team_id, kind, isPublic],
  );

  return (
    <div
      className="fixed inset-0 z-50 flex items-end justify-center bg-ink/40 px-3 pb-3 backdrop-blur-sm sm:items-center sm:p-4"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div className="w-full max-w-md rounded-3xl bg-white p-5 shadow-2xl">
        <div className="mb-3 flex items-start justify-between gap-3">
          <div className="min-w-0">
            <p className="section-label text-brand-700">
              Riwayat {kind === "tyfcb" ? "TYFCB" : "Visitor"}
            </p>
            <h3 className="text-lg font-black text-ink">{team.nama_tim}</h3>
          </div>
          <button
            type="button"
            onClick={onClose}
            aria-label="Tutup"
            className="grid h-10 w-10 shrink-0 place-items-center rounded-full text-muted transition hover:bg-brand-50 hover:text-ink"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="max-h-[60vh] overflow-y-auto">
          {loading && <Spinner />}
          {error && <ErrorNote message={error} />}
          {data && data.length === 0 && <EmptyState message="Belum ada riwayat." />}
          {data && data.length > 0 && (
            <ul className="divide-y divide-brand-50">
              {data.map((entry) => (
                <li key={entry.id} className="flex items-center justify-between gap-3 py-2.5">
                  <div className="min-w-0">
                    <p className="truncate text-sm font-semibold text-ink">
                      {entry.keterangan ?? "—"}
                    </p>
                    <p className="text-xs text-muted">{formatDate(entry.created_at)}</p>
                  </div>
                  <span
                    className={`num shrink-0 text-sm font-black ${
                      entry.points >= 0 ? "text-emerald-600" : "text-red-600"
                    }`}
                  >
                    {entry.points >= 0 ? "+" : ""}
                    {entry.points}
                  </span>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>
    </div>
  );
}
