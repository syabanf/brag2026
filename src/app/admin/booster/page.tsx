import { AppShell } from "@/components/app-shell";
import { requireUser } from "@/lib/auth";
import { query } from "@/lib/db";
import { BoosterClient, type BoosterRow } from "./booster-client";

async function getBoosters(): Promise<BoosterRow[]> {
  const { rows } = await query<BoosterRow>(`
    select
      be.id,
      be.judul,
      be.deskripsi,
      to_char(be.tanggal_mulai,    'YYYY-MM-DD') as tanggal_mulai,
      to_char(be.tanggal_berakhir, 'YYYY-MM-DD') as tanggal_berakhir,
      be.poin,
      be.multiplier::float8 as multiplier,
      be.band_min::float8   as band_min,
      be.band_max::float8   as band_max,
      be.status
    from booster_events be
    join event_seasons es on es.id = be.season_id
    where es.nama = 'BRAG 2026'
    order by be.tanggal_mulai desc
  `, []);
  return rows;
}

export default async function AdminBoosterPage() {
  const { user } = await requireUser();
  if (user.role !== "admin" && user.role !== "super_admin") {
    return <div className="p-8 text-center text-muted">Akses ditolak.</div>;
  }

  const boosters = await getBoosters();

  return (
    <AppShell>
      <div className="mb-6">
        <p className="text-sm font-semibold uppercase tracking-[0.14em] text-brand-700">Admin Area</p>
        <h1 className="mt-2 text-3xl font-black text-ink">Kelola Booster</h1>
        <p className="mt-1 text-muted">Buat dan atur event booster point mingguan.</p>
      </div>
      <BoosterClient initial={boosters} />
    </AppShell>
  );
}
