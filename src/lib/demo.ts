export const DEMO_COOKIE = "brag_demo";
export const DEMO_ROLE_COOKIE = "brag_demo_role";

export type DemoRole = "admin" | "captain" | "member";

// Each demo persona maps onto a real seeded account so every join, team
// lookup and score aggregation resolves exactly as it would in production.
export const DEMO_ACCOUNTS: Record<DemoRole, { email: string; label: string; blurb: string }> = {
  admin:   { email: "ilham@wit.id",    label: "Admin",   blurb: "Growth Coordinator — verifikasi & master data" },
  captain: { email: "m11@brag2026.id", label: "Captain", blurb: "Kapten Tim 1 — input atas nama anggota" },
  member:  { email: "m25@brag2026.id", label: "Member",  blurb: "Anggota Tim 2 — catat kontribusi sendiri" }
};

export const DEMO_ROLES = Object.keys(DEMO_ACCOUNTS) as DemoRole[];

/**
 * Whether the demo option is offered at all. Set NEXT_PUBLIC_DEMO_MODE=false
 * to hide it entirely (production).
 */
export function isDemoAvailable() {
  return process.env.NEXT_PUBLIC_DEMO_MODE !== "false";
}

export function isDemoRole(value: string | undefined | null): value is DemoRole {
  return value === "admin" || value === "captain" || value === "member";
}
