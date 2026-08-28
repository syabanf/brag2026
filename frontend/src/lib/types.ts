export type Role = "member" | "captain" | "admin";
export type ColorStatus = "merah" | "kuning" | "hijau";
export type TyfcbStatus = "pending" | "verified" | "rejected" | "void";
export type VisitorStatus = "terdaftar" | "hadir" | "hadir_penuh";

export type User = {
  id: string;
  email: string;
  full_name: string;
  role: Role;
};

export type Season = {
  id: string;
  nama: string;
  starts_on: string | null;
  ends_on: string | null;
  status: string;
};

export type Team = {
  id: string;
  season_id: string;
  nama_tim: string;
  member_count?: number;
};

export type Classification = {
  id: string;
  nama: string;
};

export type Member = {
  id: string;
  user_id: string;
  season_id: string;
  team_id: string | null;
  klasifikasi_id: string | null;
  color_status: ColorStatus;
  is_active: boolean;
  full_name: string;
  email: string;
  role: Role;
  nama_tim: string | null;
  klasifikasi_nama: string | null;
};

export type TyfcbEntry = {
  id: string;
  season_id: string;
  giver_id: string;
  receiver_id: string;
  nilai: number;
  tanggal: string;
  status: TyfcbStatus;
  computed_score: number | null;
  pair_ordinal: number | null;
  event_multiplier_applied: number | null;
  rejection_reason: string | null;
  giver_name?: string;
  receiver_name?: string;
  created_at: string;
};

export type Visitor = {
  id: string;
  season_id: string;
  nama: string;
  kontak: string;
  inviter_id: string;
  tanggal_undang: string;
  status_hadir: VisitorStatus;
  is_converted: boolean;
  is_void: boolean;
  tanggal_konversi: string | null;
  inviter_name?: string;
  nama_tim?: string | null;
  created_at: string;
};

export type LedgerEntry = {
  id: string;
  season_id: string;
  member_id: string | null;
  team_id: string | null;
  kategori: "tyfcb" | "visitor" | "bonus";
  points: number;
  sumber_ref: string | null;
  keterangan: string | null;
  created_at: string;
};

export type BoosterEvent = {
  id: string;
  season_id: string;
  judul: string;
  deskripsi: string | null;
  tanggal_mulai: string;
  tanggal_berakhir: string;
  poin: number;
  status: string;
};

export type Badge = {
  badge_code: string;
  nama: string;
  deskripsi: string;
  ikon: string | null;
  earned_at?: string | null;
};

export type TeamScore = {
  team_id: string;
  nama_tim: string;
  score_overall: number;
  score_tyfcb: number;
  score_visitor: number;
  score_bonus: number;
  nilai_tyfcb: number;
  count_tyfcb: number;
  count_visitor: number;
};

export type MemberScore = {
  member_id: string;
  full_name: string;
  nama_tim: string | null;
  score_overall: number;
  score_tyfcb: number;
  score_visitor: number;
  score_bonus: number;
};

export type Dashboard = {
  season: Season | null;
  member: Member | null;
  member_score: MemberScore | null;
  teams: TeamScore[];
  my_team: TeamScore | null;
  active_boosters: BoosterEvent[];
  recent_tyfcb: TyfcbEntry[];
  badges: Badge[];
  total_tyfcb_tx: number;
  total_tyfcb_idr: number;
  total_visitor: number;
};

export type Leaderboard = {
  teams: TeamScore[];
  members: MemberScore[];
};

export type CaptainTeam = {
  team_id: string;
  members: Member[];
  pending_tyfcb: TyfcbEntry[];
  terdaftar_visitors: Visitor[];
};
