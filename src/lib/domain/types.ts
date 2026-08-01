// BRAG Competition – Domain Types
// Aligned with Spesifikasi Aplikasi v1.0

// ─── Auth ───────────────────────────────────────────────────────────────────

export type AppRole = 'member' | 'admin';

// ─── Competition Enums ───────────────────────────────────────────────────────

export type ColorStatus   = 'merah' | 'kuning' | 'hijau';
export type TyfcbStatus   = 'pending' | 'verified' | 'rejected';
export type VisitorStatus = 'terdaftar' | 'hadir' | 'hadir_penuh';
export type LedgerKategori = 'tyfcb' | 'visitor' | 'bonus';
export type PrizeAlokasi  = 'kategori' | 'undian';
export type PrizeStatus   = 'pending' | 'approved' | 'rejected' | 'awarded';
export type RaffleSumber  = 'score' | 'visitor' | 'tyfcb_pair';
export type SeasonStatus  = 'draft' | 'active' | 'completed';

// ─── Core Entities ──────────────────────────────────────────────────────────

export type EventSeason = {
  id: string;
  nama: string;
  starts_on: string | null;
  ends_on: string | null;
  status: SeasonStatus;
  created_at: string;
};

export type Classification = {
  id: string;
  nama: string;
};

export type Team = {
  id: string;
  season_id: string;
  nama_tim: string;
};

/** Competition profile for one member in one season */
export type Member = {
  id: string;
  user_id: string;
  season_id: string;
  team_id: string | null;
  klasifikasi_id: string | null;
  color_status: ColorStatus;
  is_active: boolean;
  created_at: string;
  // joined from app_users
  full_name: string;
  email: string;
  role: AppRole;
  // joined from classifications
  klasifikasi_nama?: string;
  // joined from teams
  nama_tim?: string;
};

// ─── TYFCB ──────────────────────────────────────────────────────────────────

export type TyfcbEntry = {
  id: string;
  season_id: string;
  giver_id: string;
  receiver_id: string;
  nilai: number;
  tanggal: string;
  bukti_url: string | null;
  status: TyfcbStatus;
  computed_score: number | null;
  pair_ordinal: number | null;
  event_multiplier_applied: number | null;
  rejection_reason: string | null;
  verified_by: string | null;
  verified_at: string | null;
  created_at: string;
  // joined
  giver_name?: string;
  receiver_name?: string;
};

// ─── Scoring Engine ──────────────────────────────────────────────────────────

/** Band score lookup (spec §4.1 Step 1) */
export const TYFCB_BANDS: { min: number; max: number; score: number }[] = [
  { min: 0,           max: 500_000,       score: 10  },
  { min: 500_000,     max: 2_000_000,     score: 25  },
  { min: 2_000_000,   max: 10_000_000,    score: 50  },
  { min: 10_000_000,  max: 50_000_000,    score: 80  },
  { min: 50_000_000,  max: 250_000_000,   score: 120 },
  { min: 250_000_000, max: 500_000_000,   score: 150 },
  { min: 500_000_000, max: Infinity,      score: 200 },
];

/** Pair ordinal penalty lookup (spec §4.1 Step 2) */
export const PAIR_PENALTIES: { maxOrdinal: number; penalty: number }[] = [
  { maxOrdinal: 2,        penalty: 1.0 },
  { maxOrdinal: 5,        penalty: 0.7 },
  { maxOrdinal: Infinity, penalty: 0.5 },
];

/** Visitor milestone increments (spec §4.2) */
export const VISITOR_MILESTONES = {
  hadir:       20,  // terdaftar → hadir
  hadir_penuh: 30,  // hadir → hadir_penuh (cumulative: 50)
  converted:   100, // hadir_penuh → converted (cumulative: 150)
} as const;

/** Flat bonuses — NOT multiplied (spec §4.3) */
export const FLAT_BONUSES = {
  HIGH_ROLLER: 50,
  ONE_TO_ONE:  30,
  STREAK:      40,
} as const;

/** Team bonuses (spec §4.4) */
export const TEAM_BONUSES = {
  merah_to_kuning: 75,
  kuning_to_hijau: 150,
  full_roster:     100,
} as const;

// ─── Weekly Events ───────────────────────────────────────────────────────────

export type WeeklyEventCode =
  | 'CAT_CAROUSEL'
  | 'VISITOR_BLITZ'
  | 'CLOSING_WEEK'
  | 'SPREAD_LOVE'
  | 'UNDERDOG'
  | 'POWER_TEAM'
  | 'HIGH_ROLLER'
  | 'NEW_BLOOD'
  | 'ONE_TO_ONE'
  | 'DOUBLE_UP'
  | 'STREAK'
  | 'FOUNDER';

export type WeeklyEvent = {
  id: string;
  season_id: string;
  minggu_ke: number;
  event_code: WeeklyEventCode;
  target_classification_id: string | null;
  tanggal_mulai: string;
  tanggal_selesai: string;
  created_at: string;
  // joined
  target_classification_nama?: string;
};

export const WEEKLY_EVENT_BANK: Record<WeeklyEventCode, { nama: string; mekanik: string }> = {
  CAT_CAROUSEL:  { nama: 'Category Carousel',   mekanik: 'TYFCB ke member klasifikasi terpilih = 2×' },
  VISITOR_BLITZ: { nama: 'Visitor Blitz',        mekanik: 'Semua score visitor = 1.5×' },
  CLOSING_WEEK:  { nama: 'Closing Week',         mekanik: 'Konversi member = 2×' },
  SPREAD_LOVE:   { nama: 'Spread the Love',      mekanik: 'TYFCB ke receiver baru (pair_ordinal=1) = 2×' },
  UNDERDOG:      { nama: 'Underdog Week',        mekanik: 'TYFCB ke bisnis member status merah/kuning = 2×' },
  POWER_TEAM:    { nama: 'Power Team Week',      mekanik: 'TYFCB dalam contact sphere/power team = 1.5×' },
  HIGH_ROLLER:   { nama: 'High Roller Day',      mekanik: 'TYFCB tunggal tertinggi hari itu = flat +50' },
  NEW_BLOOD:     { nama: 'New Blood',            mekanik: 'Tiap visitor mencapai milestone hadir = 2×' },
  ONE_TO_ONE:    { nama: '1-2-1 Payoff',         mekanik: '1-2-1 berujung TYFCB minggu sama = flat +30 (admin tandai)' },
  DOUBLE_UP:     { nama: 'Double-Up Weekend',    mekanik: 'Semua score di Sabtu-Minggu = 1.5×' },
  STREAK:        { nama: 'Streak Week',          mekanik: 'Log aktivitas 3+ hari berbeda = flat +40' },
  FOUNDER:       { nama: "Founder's Frenzy",     mekanik: 'Semua score minggu itu = 1.5×' },
};

// ─── Score Ledger ────────────────────────────────────────────────────────────

export type ScoreLedgerEntry = {
  id: string;
  season_id: string;
  member_id: string | null;
  team_id: string | null;
  kategori: LedgerKategori;
  points: number;
  sumber_ref: string | null;
  keterangan: string | null;
  created_at: string;
};

// ─── Leaderboard Aggregations ────────────────────────────────────────────────

export type MemberScore = {
  member_id: string;
  full_name: string;
  team_id: string | null;
  nama_tim: string | null;
  color_status: ColorStatus;
  score_tyfcb: number;
  score_visitor: number;
  score_bonus: number;
  score_overall: number;
};

export type TeamScore = {
  team_id: string;
  nama_tim: string;
  score_tyfcb: number;
  score_visitor: number;
  score_bonus: number;
  score_overall: number;
};

// ─── Visitors ────────────────────────────────────────────────────────────────

export type Visitor = {
  id: string;
  season_id: string;
  nama: string;
  kontak: string;
  inviter_id: string;
  tanggal_undang: string;
  status_hadir: VisitorStatus;
  is_converted: boolean;
  tanggal_konversi: string | null;
  created_at: string;
  // joined
  inviter_name?: string;
};

// ─── Badges ──────────────────────────────────────────────────────────────────

export type BadgeCode =
  | 'FIRST_BLOOD' | 'HOST' | 'CLOSER' | 'CONNECTOR' | 'SPREADER'
  | 'CENTURION' | 'HAT_TRICK' | 'HIGH_ROLLER' | 'STREAK_MASTER'
  | 'TEAM_PLAYER' | 'LEVEL_UP' | 'PATRON';

export type Badge = {
  badge_code: BadgeCode;
  nama: string;
  deskripsi: string;
  ikon: string;
};

export type MemberBadge = {
  id: string;
  member_id: string;
  badge_code: BadgeCode;
  earned_at: string;
  // joined
  nama?: string;
  ikon?: string;
};

// ─── Master Data: Klasifikasi Bisnis ─────────────────────────────────────────

export type ClassificationRow = {
  id: string;
  nama: string;
  jumlah_member: number;  // how many members reference it; blocks deletion when > 0
};

// ─── Leaderboard Team History ────────────────────────────────────────────────

// Attribution mirrors the leaderboard aggregation: a TYFCB entry belongs to the
// buyer's team (giver_id), which is the side that earns the points.
export type TeamTyfcbHistoryRow = {
  id: string;
  buyer_name: string;   // giver_id — team member who earns the points
  seller_name: string;  // receiver_id — counterparty who closed the business
  nilai: number;
  tanggal: string;
  computed_score: number | null;
};

export type TeamVisitorHistoryRow = {
  id: string;
  nama: string;
  inviter_name: string;
  tanggal_undang: string;
  status_hadir: VisitorStatus;
  is_converted: boolean;
};

export type TeamHistoryResponse = {
  nama_tim: string;
  tyfcb: TeamTyfcbHistoryRow[];
  visitors: TeamVisitorHistoryRow[];
};

// ─── Prize Pool & Raffle ─────────────────────────────────────────────────────

export type PrizePool = {
  id: string;
  season_id: string;
  nama_hadiah: string;
  deskripsi: string | null;
  nilai_estimasi: number | null;
  donatur_id: string | null;
  alokasi: PrizeAlokasi;
  kategori_target: string | null;
  status: PrizeStatus;
  pemenang_id: string | null;
  created_at: string;
  // joined
  donatur_name?: string;
};

export type RaffleTicket = {
  id: string;
  season_id: string;
  member_id: string;
  sumber: RaffleSumber;
  created_at: string;
};
