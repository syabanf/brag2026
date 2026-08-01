import { AppShell } from "@/components/app-shell";
import { requireAdmin } from "@/lib/auth";
import { query } from "@/lib/db";
import { VisitorsAdminClient, type VisitorRow } from "./visitors-client";

async function getVisitors(): Promise<VisitorRow[]> {
  const { rows } = await query<VisitorRow>(`
    select
      v.id,
      u_inviter.full_name as inviter_name,
      v.nama              as visitor_name,
      v.kontak,
      to_char(v.tanggal_undang, 'DD Mon YYYY') as tanggal_undang,
      v.status_hadir::text as status_hadir,
      v.is_converted,
      v.is_void
    from visitors v
    join members m_inviter   on m_inviter.id  = v.inviter_id
    join app_users u_inviter on u_inviter.id  = m_inviter.user_id
    order by v.created_at desc
  `, []);
  return rows;
}

export default async function AdminVisitorsPage() {
  await requireAdmin();

  const visitors = await getVisitors();

  return (
    <AppShell>
      <div className="mb-6">
        <p className="text-sm font-semibold uppercase tracking-[0.14em] text-brand-700">Admin Area</p>
        <h1 className="mt-2 text-3xl font-black text-ink">Kelola Visitor</h1>
        <p className="mt-1 text-muted">Update status hadir dan konversi visitor dari member. Salah input bisa dikoreksi — poin ikut disesuaikan.</p>
      </div>
      <VisitorsAdminClient initial={visitors} />
    </AppShell>
  );
}
