import { AppShell } from "@/components/app-shell";
import { requireAdmin } from "@/lib/auth";
import { query } from "@/lib/db";
import { TyfcbAdminClient, type TyfcbRow } from "./tyfcb-client";

async function getTyfcbEntries(): Promise<TyfcbRow[]> {
  const { rows } = await query<TyfcbRow>(`
    select
      te.id,
      u_giver.full_name    as giver_name,
      u_receiver.full_name as receiver_name,
      te.nilai::text       as nilai,
      to_char(te.tanggal, 'DD Mon YYYY') as tanggal,
      te.status::text      as status,
      te.computed_score,
      te.event_multiplier_applied::text as event_multiplier_applied,
      boost.judul          as booster_judul
    from tyfcb_entries te
    join members m_giver     on m_giver.id     = te.giver_id
    join app_users u_giver   on u_giver.id     = m_giver.user_id
    join members m_receiver  on m_receiver.id  = te.receiver_id
    join app_users u_receiver on u_receiver.id = m_receiver.user_id
    -- Entries record the multiplier but not which booster produced it, so the title is
    -- re-derived with getBoosterMultiplier's rules. It returns null once an admin edits
    -- the booster out of range, which the badge falls back on gracefully.
    left join lateral (
      select be.judul
      from booster_events be
      where be.season_id = te.season_id
        and be.status = 'aktif'
        and be.multiplier > 1
        and te.tanggal between be.tanggal_mulai and be.tanggal_berakhir
        and (be.band_min is null or te.nilai >= be.band_min)
        and (be.band_max is null or te.nilai <  be.band_max)
      order by be.multiplier desc, be.created_at desc
      limit 1
    ) boost on true
    order by te.created_at desc
  `, []);
  return rows;
}

export default async function AdminTyfcbPage() {
  await requireAdmin();

  const entries = await getTyfcbEntries();

  return (
    <AppShell>
      <div className="mb-6">
        <p className="text-sm font-semibold uppercase tracking-[0.14em] text-brand-700">Admin Area</p>
        <h1 className="mt-2 text-3xl font-black text-ink">Verifikasi TYFCB</h1>
        <p className="mt-1 text-muted">Approve atau reject submission TYFCB dari member.</p>
      </div>
      <TyfcbAdminClient initial={entries} />
    </AppShell>
  );
}
