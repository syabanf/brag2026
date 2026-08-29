import { ShieldCheck, Users, User, type LucideIcon } from "lucide-react";

/**
 * The three personas the seed creates, offered as one-click sign-ins so a
 * demo does not open with someone typing an address from a README.
 *
 * These addresses are a contract with backend/seeds/demo.sql — change one and
 * change the other. The password is shared and deliberately obvious; these
 * accounts only ever exist in a seeded development database.
 */
export type DemoAccount = {
  email: string;
  password: string;
  label: string;
  /** What this persona can do that the others cannot. */
  blurb: string;
  icon: LucideIcon;
};

export const DEMO_PASSWORD = "demo123";

export const DEMO_ACCOUNTS: DemoAccount[] = [
  {
    email: "demo.admin@brag2026.id",
    password: DEMO_PASSWORD,
    label: "Admin",
    blurb: "Verifikasi, master data, undian",
    icon: ShieldCheck,
  },
  {
    email: "demo.captain@brag2026.id",
    password: DEMO_PASSWORD,
    label: "Captain",
    blurb: "Input atas nama Tim 1",
    icon: Users,
  },
  {
    email: "demo.member@brag2026.id",
    password: DEMO_PASSWORD,
    label: "Member",
    blurb: "Anggota biasa di Tim 2",
    icon: User,
  },
];

/**
 * Quick sign-in is for seeded databases only. It is on by default while
 * developing, and otherwise has to be asked for — a real deployment must not
 * advertise credentials on its login screen, even ones that do not work there.
 */
export const demoSignInEnabled =
  import.meta.env.VITE_DEMO_SIGNIN === "true" ||
  (import.meta.env.DEV && import.meta.env.VITE_DEMO_SIGNIN !== "false");
