"use client";

import { useState } from "react";
import {
  TeamHistoryDialog,
  type HistoryKind,
} from "@/components/team-history-dialog";

type Props = {
  teamId: string;
  teamName: string;
  nilaiTyfcb: number;
  countTyfcb: number;
  countVisitor: number;
};

function useHistory() {
  const [kind, setKind] = useState<HistoryKind | null>(null);
  return { kind, open: setKind, close: () => setKind(null) };
}

/** Top-3 cards: two tappable chips side by side. */
export function TeamHistoryChips({
  teamId,
  teamName,
  nilaiTyfcb,
  countTyfcb,
  countVisitor,
  chipClassName,
}: Props & { chipClassName: string }) {
  const { kind, open, close } = useHistory();

  return (
    <>
      <div className="mt-4 grid grid-cols-2 gap-2 text-center text-xs font-bold">
        <button
          type="button"
          onClick={() => open("tyfcb")}
          aria-label={`Lihat riwayat TYFCB ${teamName}`}
          className={`flex flex-col justify-center rounded-xl px-2 py-2 transition hover:brightness-105 active:scale-[0.97] ${chipClassName}`}
        >
          <span>TYFCB Rp {nilaiTyfcb.toLocaleString("id-ID")}</span>
          <span className="opacity-70">{countTyfcb}× transaksi</span>
        </button>
        <button
          type="button"
          onClick={() => open("visitor")}
          aria-label={`Lihat riwayat Visitor ${teamName}`}
          className={`flex flex-col justify-center rounded-xl px-2 py-2 transition hover:brightness-105 active:scale-[0.97] ${chipClassName}`}
        >
          Visitor {countVisitor}
        </button>
      </div>

      {kind && (
        <TeamHistoryDialog
          teamId={teamId}
          teamName={teamName}
          kind={kind}
          scope="public"
          onClose={close}
        />
      )}
    </>
  );
}

/** Rank 4+ rows: the same two targets, rendered inline in the subtitle line. */
export function TeamHistoryInline({
  teamId,
  teamName,
  nilaiTyfcb,
  countTyfcb,
  countVisitor,
}: Props) {
  const { kind, open, close } = useHistory();
  const linkCls =
    "underline decoration-dotted underline-offset-2 transition hover:text-brand-600";

  return (
    <>
      <p className="text-xs text-gray-400">
        <button
          type="button"
          onClick={() => open("tyfcb")}
          aria-label={`Lihat riwayat TYFCB ${teamName}`}
          className={linkCls}
        >
          TYFCB Rp {nilaiTyfcb.toLocaleString("id-ID")} ({countTyfcb}×)
        </button>
        {" · "}
        <button
          type="button"
          onClick={() => open("visitor")}
          aria-label={`Lihat riwayat Visitor ${teamName}`}
          className={linkCls}
        >
          Visitor {countVisitor}
        </button>
      </p>

      {kind && (
        <TeamHistoryDialog
          teamId={teamId}
          teamName={teamName}
          kind={kind}
          scope="public"
          onClose={close}
        />
      )}
    </>
  );
}
