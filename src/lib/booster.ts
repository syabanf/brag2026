import { TYFCB_BANDS } from "@/lib/domain/types";

export type BoosterBandRange = {
  /** Inclusive lower bound in rupiah. Null = unbounded below. */
  band_min: number | null;
  /** Exclusive upper bound in rupiah. Null = unbounded above. */
  band_max: number | null;
};

export type BoosterEffect = BoosterBandRange & {
  multiplier: number;
  poin: number;
};

/** "2 jt", "500 rb", "1 M" — compact enough for a band label on mobile. */
export function formatRupiahShort(nilai: number): string {
  if (nilai >= 1_000_000_000) return `${trimZero(nilai / 1_000_000_000)} M`;
  if (nilai >= 1_000_000) return `${trimZero(nilai / 1_000_000)} jt`;
  if (nilai >= 1_000) return `${trimZero(nilai / 1_000)} rb`;
  return String(nilai);
}

function trimZero(n: number): string {
  return String(Number(n.toFixed(2)));
}

/**
 * Band choices offered when creating a booster, mirroring TYFCB_BANDS so an admin
 * cannot define a range that straddles two scoring bands.
 */
export const BOOSTER_BAND_PRESETS: { label: string; band_min: number | null; band_max: number | null }[] = [
  { label: "Semua nilai TYFCB", band_min: null, band_max: null },
  ...TYFCB_BANDS.map((b) => ({
    label: `${formatRupiahShort(b.min)} – ${b.max === Infinity ? "ke atas" : formatRupiahShort(b.max)} (${b.score} pts)`,
    band_min: b.min,
    band_max: Number.isFinite(b.max) ? b.max : null,
  })),
];

export function boosterBandLabel({ band_min, band_max }: BoosterBandRange): string {
  if (band_min === null && band_max === null) return "Semua nilai TYFCB";
  if (band_min === null) return `< ${formatRupiahShort(band_max as number)}`;
  if (band_max === null) return `≥ ${formatRupiahShort(band_min)}`;
  return `${formatRupiahShort(band_min)} – ${formatRupiahShort(band_max)}`;
}

/**
 * What the member actually gets. A multiplier booster is scored automatically; a
 * `poin` booster is a flat announcement an admin still awards by hand, so the two
 * are labelled differently on purpose.
 */
export function boosterEffectLabel({ multiplier, poin }: Pick<BoosterEffect, "multiplier" | "poin">): string {
  if (multiplier > 1) return `${trimZero(multiplier)}x poin`;
  return `+${poin} pts`;
}

/** Preview of a band's score under a booster, for the admin form. */
export function boosterScorePreview(multiplier: number, { band_min, band_max }: BoosterBandRange): string | null {
  if (multiplier <= 1) return null;
  const band = TYFCB_BANDS.find(
    (b) => b.min === (band_min ?? 0) && (Number.isFinite(b.max) ? b.max : null) === band_max
  );
  if (!band) return null;
  return `${band.score} → ${Math.round(band.score * multiplier)} poin`;
}
