import { TYFCB_BANDS } from "@/lib/domain/types";

export type ParsedBoosterScoring = {
  multiplier: number;
  band_min: number | null;
  band_max: number | null;
};

const MAX_MULTIPLIER = 10;

/**
 * Validates the scoring half of a booster before it reaches the DB.
 *
 * A booster that multiplies must land on a real TYFCB band (or cover every band),
 * otherwise a range like 3jt–8jt would silently boost only part of the 2jt–10jt band
 * and two members with the same band would score differently.
 */
export function parseBoosterScoring(input: {
  multiplier?: unknown;
  band_min?: unknown;
  band_max?: unknown;
}): ParsedBoosterScoring | { error: string } {
  const multiplier = input.multiplier === undefined || input.multiplier === null || input.multiplier === ""
    ? 1
    : Number(input.multiplier);

  if (!Number.isFinite(multiplier) || multiplier < 1 || multiplier > MAX_MULTIPLIER) {
    return { error: `Multiplier harus antara 1 dan ${MAX_MULTIPLIER}.` };
  }

  const band_min = toBound(input.band_min);
  const band_max = toBound(input.band_max);

  if (band_min === undefined || band_max === undefined) {
    return { error: "Batas band tidak valid." };
  }
  if (band_min !== null && band_max !== null && band_max <= band_min) {
    return { error: "Batas atas band harus lebih besar dari batas bawah." };
  }

  const bounded = band_min !== null || band_max !== null;
  if (bounded && !matchesKnownBand(band_min, band_max)) {
    return { error: "Rentang band harus sama persis dengan salah satu band TYFCB." };
  }

  return { multiplier, band_min, band_max };
}

function toBound(value: unknown): number | null | undefined {
  if (value === undefined || value === null || value === "") return null;
  const n = Number(value);
  if (!Number.isFinite(n) || n < 0) return undefined;
  return n;
}

function matchesKnownBand(band_min: number | null, band_max: number | null): boolean {
  return TYFCB_BANDS.some(
    (b) => b.min === (band_min ?? 0) && (Number.isFinite(b.max) ? b.max : null) === band_max
  );
}
