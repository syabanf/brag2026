import { AppShell } from "@/components/app-shell";
import { requireUser } from "@/lib/auth";
import { isAdminRole, listClassifications } from "@/lib/domain/classifications";
import { ClassificationsClient } from "./classifications-client";

export default async function AdminClassificationsPage() {
  const { user } = await requireUser();
  if (!isAdminRole(user.role)) {
    return <div className="p-8 text-center text-muted">Akses ditolak.</div>;
  }

  const classifications = await listClassifications();

  return (
    <AppShell>
      <div className="mb-6">
        <p className="text-sm font-semibold uppercase tracking-[0.14em] text-brand-700">Admin Area</p>
        <h1 className="mt-2 text-3xl font-black text-ink">Klasifikasi Bisnis</h1>
        <p className="mt-1 text-muted">
          Master data kategori bisnis member. Klasifikasi yang sedang dipakai member
          tidak bisa dihapus — ubah dulu klasifikasi member tersebut.
        </p>
      </div>
      <ClassificationsClient initial={classifications} />
    </AppShell>
  );
}
