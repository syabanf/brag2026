import { UserPlus } from "lucide-react";
import { AppShell } from "@/components/app-shell";
import { requireAdmin } from "@/lib/auth";
import { query } from "@/lib/db";
import { MemberTable } from "./members-client";

async function getMembers() {
  const { rows } = await query<{
    id: string;
    user_id: string;
    full_name: string;
    email: string;
    role: string;
    team_id: string | null;
    nama_tim: string | null;
    klasifikasi_id: string | null;
    klasifikasi_nama: string | null;
    color_status: string;
    is_active: boolean;
  }>(`
    select
      m.id, m.user_id, m.team_id, m.klasifikasi_id,
      m.color_status, m.is_active,
      u.full_name, u.email, u.role,
      t.nama_tim,
      c.nama as klasifikasi_nama
    from members m
    join app_users u on u.id = m.user_id
    join event_seasons es on es.id = m.season_id
    left join teams t on t.id = m.team_id
    left join classifications c on c.id = m.klasifikasi_id
    where es.nama = 'BRAG 2026'
    order by
      substring(t.nama_tim, 5)::int,
      u.full_name
  `);
  return rows;
}

async function getTeams() {
  const { rows } = await query<{ id: string; nama_tim: string }>(`
    select t.id, t.nama_tim
    from teams t
    join event_seasons es on es.id = t.season_id
    where es.nama = 'BRAG 2026'
    order by substring(t.nama_tim, 5)::int
  `);
  return rows;
}

async function getKlasifikasi() {
  const { rows } = await query<{ id: string; nama: string }>(
    `select id, nama from classifications order by nama`
  );
  return rows;
}

export default async function AdminMembersPage() {
  await requireAdmin();
  const [members, teams, klasifikasi] = await Promise.all([
    getMembers(),
    getTeams(),
    getKlasifikasi(),
  ]);

  return (
    <AppShell>
      <div className="mb-5 flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <h1 className="text-2xl font-black leading-tight tracking-tight text-ink sm:text-3xl">
            Kelola Member
          </h1>
          <p className="mt-1 text-sm text-muted">
            {members.length} member · 10 tim · BRAG 2026
          </p>
        </div>
        <a
          href="/admin/members/new"
          className="flex min-h-11 shrink-0 items-center gap-2 rounded-full bg-brand-600 px-4 text-sm font-black text-white shadow transition hover:bg-brand-700 sm:px-5"
        >
          <UserPlus className="h-4 w-4" />
          Tambah
          <span className="hidden sm:inline">Member</span>
        </a>
      </div>

      <MemberTable members={members} teams={teams} klasifikasi={klasifikasi} />
    </AppShell>
  );
}
