import { Shield } from "lucide-react";
import { AppShell } from "@/components/app-shell";
import { requireCaptain } from "@/lib/auth";
import { query } from "@/lib/db";
import { getCaptainContext } from "@/lib/captain";
import { CaptainPanel } from "./captain-client";

export type TeamMember = {
  id: string;
  full_name: string;
  email: string;
  color_status: string;
  is_active: boolean;
};

export type SeasonMember = {
  id: string;
  full_name: string;
  nama_tim: string | null;
};

export type PendingTyfcb = {
  id: string;
  seller_id: string;   // receiver_id in DB — team member who submitted
  buyer_name: string;  // giver_id in DB — buyer who gets TYFCB points
  nilai: number;
  tanggal: string;
  computed_score: number;
};

export type TerdaftarVisitor = {
  id: string;
  inviter_id: string;
  nama: string;
  kontak: string;
  tanggal_undang: string;
};

async function getTeamName(teamId: string) {
  const { rows } = await query<{ nama_tim: string }>(
    `select nama_tim from teams where id = $1 limit 1`,
    [teamId]
  );
  return rows[0]?.nama_tim ?? "Tim Kamu";
}

async function getTeamMembers(teamId: string, seasonId: string): Promise<TeamMember[]> {
  const { rows } = await query<TeamMember>(`
    select m.id, u.full_name, u.email, m.color_status::text as color_status, m.is_active
    from members m
    join app_users u on u.id = m.user_id
    where m.team_id = $1 and m.season_id = $2
    order by u.full_name
  `, [teamId, seasonId]);
  return rows;
}

async function getAllSeasonMembers(seasonId: string): Promise<SeasonMember[]> {
  const { rows } = await query<SeasonMember>(`
    select m.id, u.full_name, t.nama_tim
    from members m
    join app_users u on u.id = m.user_id
    left join teams t on t.id = m.team_id
    where m.season_id = $1
    order by u.full_name
  `, [seasonId]);
  return rows;
}

async function getPendingTyfcb(teamId: string, seasonId: string): Promise<PendingTyfcb[]> {
  // After the TYFCB logic fix: giver_id = buyer (gets points), receiver_id = seller (team member who submitted)
  const { rows } = await query<PendingTyfcb>(`
    select
      te.id,
      te.receiver_id   as seller_id,
      u_buyer.full_name as buyer_name,
      te.nilai::int    as nilai,
      to_char(te.tanggal, 'DD Mon YYYY') as tanggal,
      te.computed_score
    from tyfcb_entries te
    join members m_seller on m_seller.id = te.receiver_id
    join members m_buyer  on m_buyer.id  = te.giver_id
    join app_users u_buyer on u_buyer.id = m_buyer.user_id
    where m_seller.team_id = $1
      and te.season_id = $2
      and te.status = 'pending'
    order by te.created_at desc
  `, [teamId, seasonId]);
  return rows;
}

async function getTerdaftarVisitors(teamId: string, seasonId: string): Promise<TerdaftarVisitor[]> {
  const { rows } = await query<TerdaftarVisitor>(`
    select v.id, v.inviter_id, v.nama, v.kontak,
           to_char(v.tanggal_undang, 'DD Mon YYYY') as tanggal_undang
    from visitors v
    join members m on m.id = v.inviter_id
    where m.team_id = $1
      and v.season_id = $2
      and v.status_hadir = 'terdaftar'
      and v.is_void = false
    order by v.created_at desc
  `, [teamId, seasonId]);
  return rows;
}

export default async function CaptainPage() {
  const { user } = await requireCaptain();

  const ctx = await getCaptainContext(user.id);
  if (!ctx?.team_id) {
    return (
      <AppShell>
        <div className="flex flex-col items-center justify-center py-24 text-center">
          <Shield className="mb-4 h-12 w-12 text-muted" />
          <p className="text-lg font-bold text-ink">Kamu belum tergabung dalam tim.</p>
          <p className="mt-1 text-sm text-muted">Hubungi Admin untuk mendaftarkan timmu.</p>
        </div>
      </AppShell>
    );
  }

  const [teamName, members, allMembers, pendingTyfcb, terdaftarVisitors] = await Promise.all([
    getTeamName(ctx.team_id),
    getTeamMembers(ctx.team_id, ctx.season_id),
    getAllSeasonMembers(ctx.season_id),
    getPendingTyfcb(ctx.team_id, ctx.season_id),
    getTerdaftarVisitors(ctx.team_id, ctx.season_id),
  ]);

  return (
    <AppShell>
      <div className="mb-6">
        <div className="flex items-center gap-2">
          <Shield className="h-5 w-5 text-amber-600" />
          <p className="text-sm font-bold uppercase tracking-[0.14em] text-amber-700">Panel Kapten</p>
        </div>
        <h1 className="mt-1 text-3xl font-black text-ink">{teamName}</h1>
        <p className="mt-1 text-muted">{members.length} anggota · {pendingTyfcb.length} TYFCB pending · {terdaftarVisitors.length} visitor terdaftar</p>
      </div>

      <CaptainPanel
        members={members}
        allMembers={allMembers}
        pendingTyfcb={pendingTyfcb}
        terdaftarVisitors={terdaftarVisitors}
      />
    </AppShell>
  );
}
