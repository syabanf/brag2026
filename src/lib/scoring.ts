import { TYFCB_BANDS, PAIR_PENALTIES } from "@/lib/domain/types";

export type TyfcbScoreOptions = {
  /** Per-season setting from event_seasons.pair_penalty_enabled */
  pairPenaltyEnabled: boolean;
  eventMultiplier?: number;
};

/** Band score from transaction value (spec §4.1 Step 1) */
export function getBand(nilai: number): number {
  const band = TYFCB_BANDS.find((b) => nilai < b.max);
  return band ? band.score : 0;
}

/**
 * Repeat transactions between the same buyer→seller pair decay in value (spec §4.1 Step 2).
 * Returns 1.0 when the season has the penalty switched off, so every entry scores at full band.
 */
export function getPairPenalty(pairOrdinal: number, enabled: boolean): number {
  if (!enabled) return 1.0;
  const tier = PAIR_PENALTIES.find((p) => pairOrdinal <= p.maxOrdinal);
  return tier ? tier.penalty : 0.5;
}

/** computed_score = round(B × P × M) (spec §4.1) */
export function computeTyfcbScore(
  nilai: number,
  pairOrdinal: number,
  { pairPenaltyEnabled, eventMultiplier = 1.0 }: TyfcbScoreOptions
): number {
  return Math.round(
    getBand(nilai) * getPairPenalty(pairOrdinal, pairPenaltyEnabled) * eventMultiplier
  );
}
