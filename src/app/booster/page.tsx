import Link from "next/link";
import { CalendarDays, ChevronRight, Zap } from "lucide-react";
import { AppShell } from "@/components/app-shell";
import { requireUser } from "@/lib/auth";
import { query } from "@/lib/db";
import { boosterBandLabel, boosterEffectLabel } from "@/lib/booster";

type BoosterState = "running" | "coming_soon" | "inactive";

type BoosterRow = {
  id: string;
  judul: string;
  deskripsi: string | null;
  tanggal_mulai: string;
  tanggal_berakhir: string;
  poin: number;
  multiplier: number;
  band_min: number | null;
  band_max: number | null;
  state: BoosterState;
};

async function getAllBoosters(): Promise<BoosterRow[]> {
  const { rows } = await query<BoosterRow>(`
    select
      be.id,
      be.judul,
      be.deskripsi,
      to_char(be.tanggal_mulai,    'DD Mon YYYY') as tanggal_mulai,
      to_char(be.tanggal_berakhir, 'DD Mon YYYY') as tanggal_berakhir,
      be.poin,
      be.multiplier::float8 as multiplier,
      be.band_min::float8   as band_min,
      be.band_max::float8   as band_max,
      case
        when be.status = 'aktif' and current_date between be.tanggal_mulai and be.tanggal_berakhir
          then 'running'
        when be.status = 'aktif' and current_date < be.tanggal_mulai
          then 'coming_soon'
        else 'inactive'
      end as state
    from booster_events be
    join event_seasons es on es.id = be.season_id
    where es.nama = 'BRAG 2026'
    order by
      case
        when be.status = 'aktif' and current_date between be.tanggal_mulai and be.tanggal_berakhir then 0
        when be.status = 'aktif' and current_date < be.tanggal_mulai then 1
        else 2
      end,
      be.tanggal_mulai desc
  `, []);
  return rows;
}

function BoosterCard({ b }: { b: BoosterRow }) {
  const running    = b.state === "running";
  const comingSoon = b.state === "coming_soon";
  const inactive   = b.state === "inactive";

  const cardBase = "group flex items-start gap-4 overflow-hidden rounded-2xl border p-4 transition";
  const cardStyle = running
    ? `${cardBase} border-brand-300 bg-gradient-to-br from-brand-600 via-red-600 to-orange-500 text-white shadow-lift hover:brightness-110`
    : comingSoon
      ? `${cardBase} border-indigo-300 bg-gradient-to-br from-indigo-500 via-blue-600 to-violet-500 text-white shadow-sm`
      : `${cardBase} cursor-default border-gray-200 bg-gray-100`;

  const iconStyle = inactive
    ? "mt-0.5 grid h-11 w-11 shrink-0 place-items-center rounded-full bg-gray-200 text-gray-400"
    : "mt-0.5 grid h-11 w-11 shrink-0 place-items-center rounded-full bg-white/20 text-white";

  const inner = (
    <>
      <span className={iconStyle}>
        <Zap className="h-5 w-5" />
      </span>

      <div className="min-w-0 flex-1">
        {running && (
          <span className="mb-1.5 inline-block rounded-full bg-white/20 px-2.5 py-0.5 text-[0.65rem] font-black uppercase tracking-wider text-white">
            Aktif Sekarang
          </span>
        )}
        {comingSoon && (
          <span className="mb-1.5 inline-block rounded-full bg-white/20 px-2.5 py-0.5 text-[0.65rem] font-black uppercase tracking-wider text-white">
            Coming Soon
          </span>
        )}
        {inactive && (
          <span className="mb-1.5 inline-block rounded-full bg-gray-200 px-2.5 py-0.5 text-[0.65rem] font-black uppercase tracking-wider text-gray-500">
            Tidak Aktif
          </span>
        )}

        <p className={`text-base font-black leading-snug ${inactive ? "text-gray-500" : "text-white"}`}>
          {b.judul}
        </p>

        <div className="mt-1 flex flex-wrap items-center gap-1.5">
          <span className={`inline-block rounded-full px-2.5 py-0.5 text-xs font-black ${
            inactive ? "bg-gray-200 text-gray-400" : "bg-white/20 text-white"
          }`}>
            {boosterEffectLabel(b)}
          </span>
          {b.multiplier > 1 && (
            <span className={`inline-block rounded-full px-2.5 py-0.5 text-xs font-black ${
              inactive ? "bg-gray-200 text-gray-400" : "bg-white/20 text-white"
            }`}>
              TYFCB {boosterBandLabel(b)}
            </span>
          )}
        </div>

        {b.deskripsi && (
          <p className={`mt-1.5 line-clamp-2 text-sm ${inactive ? "text-gray-400" : "text-white/80"}`}>
            {b.deskripsi}
          </p>
        )}

        <div className={`mt-2 flex items-center gap-1.5 text-xs font-medium ${
          inactive ? "text-gray-400" : "text-white/70"
        }`}>
          <CalendarDays className="h-3.5 w-3.5 shrink-0" />
          {b.tanggal_mulai} — {b.tanggal_berakhir}
        </div>
      </div>

      {running && (
        <ChevronRight className="mt-1 h-5 w-5 shrink-0 text-white/60 transition group-hover:translate-x-0.5 group-hover:text-white" />
      )}
    </>
  );

  if (running || comingSoon) {
    return (
      <Link href={`/booster/${b.id}`} className={cardStyle}>
        {inner}
      </Link>
    );
  }

  return <div className={cardStyle}>{inner}</div>;
}

export default async function BoosterPage() {
  await requireUser();
  const boosters = await getAllBoosters();

  return (
    <AppShell>
      <div className="mb-6">
        <p className="text-sm font-semibold uppercase tracking-[0.14em] text-brand-700">
          Booster Point
        </p>
        <h1 className="mt-2 text-3xl font-black text-ink">Daftar Booster</h1>
        <p className="mt-2 text-muted">
          Event booster yang ditetapkan oleh Grow Coordinator.
          Raih poin tambahan selama periode berlaku.
        </p>
      </div>

      {boosters.length === 0 ? (
        <div className="flex flex-col items-center gap-4 rounded-2xl border border-dashed border-brand-200 bg-white py-20 text-center">
          <span className="grid h-16 w-16 place-items-center rounded-full bg-brand-50 text-brand-600">
            <Zap className="h-8 w-8" />
          </span>
          <div>
            <p className="text-lg font-black text-ink">Belum ada booster dijadwalkan</p>
            <p className="mt-1 text-sm text-muted">
              Grow Coordinator akan mengisi jadwal booster minggu per minggu.
            </p>
          </div>
        </div>
      ) : (
        <div className="space-y-3">
          {boosters.map((b) => (
            <BoosterCard key={b.id} b={b} />
          ))}
        </div>
      )}
    </AppShell>
  );
}
