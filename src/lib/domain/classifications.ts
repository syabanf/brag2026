import { query } from "@/lib/db";
import type { ClassificationRow } from "@/lib/domain/types";

export const NAMA_MAX_LENGTH = 60;
export const NAMA_TAKEN = "Klasifikasi dengan nama itu sudah ada.";

export function isAdminRole(role: string): boolean {
  return role === "admin" || role === "super_admin";
}

type ValidationResult =
  | { ok: true; value: string }
  | { ok: false; error: string };

export function validateNama(nama: unknown): ValidationResult {
  if (typeof nama !== "string") {
    return { ok: false, error: "Nama klasifikasi wajib diisi." };
  }

  const value = nama.trim();
  if (!value) {
    return { ok: false, error: "Nama klasifikasi wajib diisi." };
  }
  if (value.length > NAMA_MAX_LENGTH) {
    return {
      ok: false,
      error: `Nama klasifikasi maksimal ${NAMA_MAX_LENGTH} karakter.`,
    };
  }

  return { ok: true, value };
}

// jumlah_member drives the delete guard — a classification still in use may
// not be removed, since members.klasifikasi_id references it.
export async function listClassifications(): Promise<ClassificationRow[]> {
  const { rows } = await query<ClassificationRow>(`
    select c.id, c.nama, count(m.id)::int as jumlah_member
    from classifications c
    left join members m on m.klasifikasi_id = c.id
    group by c.id, c.nama
    order by c.nama
  `, []);
  return rows;
}
