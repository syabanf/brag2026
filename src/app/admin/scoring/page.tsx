import { AppShell } from "@/components/app-shell";
import { requireAdmin } from "@/lib/auth";
import { computePairPenaltyImpact, getActiveSeason } from "@/lib/scoring-settings";
import { PairPenaltyClient } from "./pair-penalty-client";

export default async function AdminScoringPage() {
  await requireAdmin();

  const season = await getActiveSeason();

  if (!season) {
    return (
      <AppShell>
        <p className="text-muted">Season aktif tidak ditemukan.</p>
      </AppShell>
    );
  }

  const impact = await computePairPenaltyImpact(season.id, !season.pair_penalty_enabled);

  return (
    <AppShell>
      <div className="mb-6">
        <p className="text-sm font-semibold uppercase tracking-[0.14em] text-brand-700">Admin Area</p>
        <h1 className="mt-2 text-3xl font-black text-ink">Aturan Skor</h1>
        <p className="mt-1 text-muted">
          Atur pair penalty untuk season {season.nama} dan lihat dampaknya sebelum diterapkan.
        </p>
      </div>
      <PairPenaltyClient initialEnabled={season.pair_penalty_enabled} initialImpact={impact} />
    </AppShell>
  );
}
