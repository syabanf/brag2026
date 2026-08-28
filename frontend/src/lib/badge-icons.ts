import {
  Award,
  Crown,
  Flame,
  Gem,
  Gift,
  Handshake,
  Home,
  Link2,
  Medal,
  Radio,
  TrendingUp,
  Trophy,
  Users,
  type LucideIcon,
} from "lucide-react";

/**
 * Badge artwork as icons rather than emoji. Emoji render differently on every
 * platform and cannot inherit colour or size, so the set never looks like one
 * system. The database still carries an `ikon` column; it is ignored here.
 */
const BADGE_ICONS: Record<string, LucideIcon> = {
  FIRST_BLOOD: Medal,
  HOST: Home,
  CLOSER: Handshake,
  CONNECTOR: Link2,
  SPREADER: Radio,
  CENTURION: Trophy,
  HAT_TRICK: Crown,
  HIGH_ROLLER: Gem,
  STREAK_MASTER: Flame,
  TEAM_PLAYER: Users,
  LEVEL_UP: TrendingUp,
  PATRON: Gift,
};

export function badgeIcon(code: string): LucideIcon {
  return BADGE_ICONS[code] ?? Award;
}

/** Tints stay within the brand family so a wall of badges reads as one set. */
const BADGE_TONES: Record<string, string> = {
  FIRST_BLOOD: "bg-red-50 text-red-600",
  HOST: "bg-sky-50 text-sky-600",
  CLOSER: "bg-emerald-50 text-emerald-600",
  CONNECTOR: "bg-violet-50 text-violet-600",
  SPREADER: "bg-violet-50 text-violet-600",
  CENTURION: "bg-amber-50 text-amber-600",
  HAT_TRICK: "bg-amber-50 text-amber-600",
  HIGH_ROLLER: "bg-sky-50 text-sky-600",
  STREAK_MASTER: "bg-orange-50 text-orange-600",
  TEAM_PLAYER: "bg-emerald-50 text-emerald-600",
  LEVEL_UP: "bg-brand-50 text-brand-600",
  PATRON: "bg-brand-50 text-brand-600",
};

export function badgeTone(code: string): string {
  return BADGE_TONES[code] ?? "bg-brand-50 text-brand-600";
}
