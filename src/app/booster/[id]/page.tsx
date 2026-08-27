import Link from "next/link";
import { CalendarDays, ChevronLeft, Zap } from "lucide-react";
import { notFound } from "next/navigation";
import { AppShell } from "@/components/app-shell";
import { requireUser } from "@/lib/auth";
import { query } from "@/lib/db";
import { boosterBandLabel, boosterEffectLabel } from "@/lib/booster";

async function getBooster(id: string) {
  const { rows } = await query<{
    id: string;
    judul: string;
    deskripsi: string | null;
    poin: number;
    multiplier: number;
    band_min: number | null;
    band_max: number | null;
    tanggal_mulai: string;
    tanggal_berakhir: string;
    is_running: boolean;
  }>(`
    select
      id, judul, deskripsi, poin,
      multiplier::float8 as multiplier,
      band_min::float8   as band_min,
      band_max::float8   as band_max,
      to_char(tanggal_mulai,    'DD Mon YYYY') as tanggal_mulai,
      to_char(tanggal_berakhir, 'DD Mon YYYY') as tanggal_berakhir,
      (current_date between tanggal_mulai and tanggal_berakhir) as is_running
    from booster_events
    where id = $1 and status = 'aktif'
    limit 1
  `, [id]);
  return rows[0] ?? null;
}

export default async function BoosterDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  await requireUser();
  const { id } = await params;
  const booster = await getBooster(id);

  if (!booster) notFound();

  return (
    <AppShell>
      <div className="mx-auto max-w-xl">
        <Link
          href="/booster"
          className="mb-6 flex items-center gap-1.5 text-sm font-bold text-brand-600 hover:underline"
        >
          <ChevronLeft className="h-4 w-4" />
          Semua Booster
        </Link>

        {/* Hero card */}
        <div className="overflow-hidden rounded-3xl bg-gradient-to-br from-brand-600 via-red-600 to-orange-500 p-6 text-white shadow-lift">
          <div className="mb-4 flex items-center gap-3">
            <span className="grid h-14 w-14 place-items-center rounded-full bg-white/20">
              <Zap className="h-7 w-7 fill-white" />
            </span>
            <div>
              {booster.is_running && (
                <span className="rounded-full bg-white/20 px-3 py-1 text-xs font-black uppercase tracking-wider">
                  Aktif Sekarang
                </span>
              )}
            </div>
          </div>
          <h1 className="text-3xl font-black leading-tight">{booster.judul}</h1>
          <p className="mt-2 text-4xl font-black text-white/90">{boosterEffectLabel(booster)}</p>
          {booster.multiplier > 1 && (
            <p className="mt-1 text-sm font-bold text-white/80">
              Untuk TYFCB {boosterBandLabel(booster)}
            </p>
          )}
          <div className="mt-4 flex items-center gap-2 text-sm text-white/70">
            <CalendarDays className="h-4 w-4" />
            {booster.tanggal_mulai} — {booster.tanggal_berakhir}
          </div>
        </div>

        {/* Deskripsi */}
        {booster.deskripsi && (
          <div className="mt-5 rounded-2xl border border-brand-100 bg-white p-5">
            <p className="mb-2 text-xs font-bold uppercase tracking-[0.14em] text-brand-700">Deskripsi</p>
            <p className="text-base leading-relaxed text-ink">{booster.deskripsi}</p>
          </div>
        )}

        {/* CTA */}
        <div className="mt-4 rounded-2xl bg-brand-50 p-4 text-sm text-muted">
          <p className="font-bold text-ink">Cara mendapatkan poin booster ini:</p>
          {booster.multiplier > 1 ? (
            <p className="mt-1">
              Catat TYFCB {boosterBandLabel(booster)} dengan tanggal transaksi di dalam periode booster.
              Poinnya otomatis dikalikan {boosterEffectLabel(booster)} saat kamu submit, dan masuk
              leaderboard setelah diverifikasi Grow Coordinator.
            </p>
          ) : (
            <p className="mt-1">
              Lakukan aktivitas TYFCB atau Visitor selama periode booster aktif.
              Poin akan dikalkulasikan oleh Grow Coordinator setelah diverifikasi.
            </p>
          )}
        </div>
      </div>
    </AppShell>
  );
}
