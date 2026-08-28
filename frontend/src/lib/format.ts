export function formatPoints(points: number) {
  return new Intl.NumberFormat("en-US").format(points);
}

export function formatCurrency(value: number) {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    maximumFractionDigits: 0,
  }).format(value);
}

/**
 * Shortens large rupiah figures for dashboard tiles, where the magnitude
 * matters more than the exact digits and a full number would overflow the
 * card. 20_015_750_000 → "Rp 20 M".
 */
export function formatCurrencyCompact(value: number) {
  const units: [number, string][] = [
    [1_000_000_000_000, "T"],
    [1_000_000_000, "M"],
    [1_000_000, "jt"],
    [1_000, "rb"],
  ];

  for (const [size, suffix] of units) {
    if (Math.abs(value) >= size) {
      const scaled = value / size;
      const digits = scaled >= 100 ? 0 : scaled >= 10 ? 1 : 2;
      return `Rp ${new Intl.NumberFormat("id-ID", {
        maximumFractionDigits: digits,
      }).format(scaled)} ${suffix}`;
    }
  }

  return `Rp ${new Intl.NumberFormat("id-ID").format(value)}`;
}

export function formatDate(value: string | null | undefined) {
  if (!value) return "—";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "—";
  return new Intl.DateTimeFormat("id-ID", {
    day: "2-digit",
    month: "short",
    year: "numeric",
  }).format(date);
}

export function initials(fullName: string | undefined) {
  if (!fullName) return "?";
  return fullName
    .split(" ")
    .map((part) => part[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();
}

/** Today as yyyy-mm-dd, for date input defaults. */
export function today() {
  return new Date().toISOString().slice(0, 10);
}
