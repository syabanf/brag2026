import { clsx, type ClassValue } from "clsx";

export function cn(...inputs: ClassValue[]) {
  return clsx(inputs);
}

export function formatPoints(points: number) {
  return new Intl.NumberFormat("en-US").format(points);
}

export function formatCurrency(value: number) {
  return new Intl.NumberFormat("id-ID", {
    currency: "IDR",
    maximumFractionDigits: 0,
    style: "currency"
  }).format(value);
}

/**
 * Shortens large rupiah figures for dashboard tiles, where the exact digits
 * matter less than the magnitude and a full number would overflow the card.
 * 20_015_750_000 → "Rp 20,02 M". Pair it with the exact value in a title.
 */
export function formatCurrencyCompact(value: number) {
  const units: [number, string][] = [
    [1_000_000_000_000, "T"],
    [1_000_000_000, "M"],
    [1_000_000, "jt"],
    [1_000, "rb"]
  ];

  for (const [size, suffix] of units) {
    if (Math.abs(value) >= size) {
      const scaled = value / size;
      const digits = scaled >= 100 ? 0 : scaled >= 10 ? 1 : 2;
      const formatted = new Intl.NumberFormat("id-ID", {
        maximumFractionDigits: digits,
        minimumFractionDigits: 0
      }).format(scaled);
      return `Rp ${formatted} ${suffix}`;
    }
  }

  return `Rp ${new Intl.NumberFormat("id-ID").format(value)}`;
}
