import { NextRequest } from "next/server";
import { requireUser } from "@/lib/auth";
import { query } from "@/lib/db";
import {
  isAdminRole,
  validateNama,
  NAMA_TAKEN,
} from "@/lib/domain/classifications";

export async function PATCH(
  req: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { user } = await requireUser();
  if (!isAdminRole(user.role)) {
    return Response.json({ error: "Forbidden" }, { status: 403 });
  }

  const { id } = await params;
  const { nama } = await req.json();
  const validation = validateNama(nama);
  if (!validation.ok) {
    return Response.json({ error: validation.error }, { status: 400 });
  }

  // Exclude self so re-saving an unchanged name is not reported as a duplicate.
  const { rows: clash } = await query<{ id: string }>(
    `select id from classifications where lower(nama) = lower($1) and id <> $2 limit 1`,
    [validation.value, id]
  );
  if (clash[0]) {
    return Response.json({ error: NAMA_TAKEN }, { status: 409 });
  }

  const { rowCount } = await query(
    `update classifications set nama = $1 where id = $2`,
    [validation.value, id]
  );
  if (!rowCount) {
    return Response.json({ error: "Klasifikasi tidak ditemukan." }, { status: 404 });
  }

  return Response.json({ ok: true });
}

export async function DELETE(
  _req: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { user } = await requireUser();
  if (!isAdminRole(user.role)) {
    return Response.json({ error: "Forbidden" }, { status: 403 });
  }

  const { id } = await params;

  const { rows } = await query<{ jumlah: number }>(
    `select count(*)::int as jumlah from members where klasifikasi_id = $1`,
    [id]
  );
  const jumlah = rows[0]?.jumlah ?? 0;
  if (jumlah > 0) {
    return Response.json(
      {
        error: `Klasifikasi masih dipakai ${jumlah} member. Pindahkan member tersebut terlebih dahulu.`,
      },
      { status: 409 }
    );
  }

  const { rowCount } = await query(`delete from classifications where id = $1`, [id]);
  if (!rowCount) {
    return Response.json({ error: "Klasifikasi tidak ditemukan." }, { status: 404 });
  }

  return Response.json({ ok: true });
}
