import { NextRequest } from "next/server";
import { requireUser } from "@/lib/auth";
import { query } from "@/lib/db";
import {
  isAdminRole,
  listClassifications,
  validateNama,
  NAMA_TAKEN,
} from "@/lib/domain/classifications";

export async function GET() {
  const { user } = await requireUser();
  if (!isAdminRole(user.role)) {
    return Response.json({ error: "Forbidden" }, { status: 403 });
  }

  return Response.json({ classifications: await listClassifications() });
}

export async function POST(req: NextRequest) {
  const { user } = await requireUser();
  if (!isAdminRole(user.role)) {
    return Response.json({ error: "Forbidden" }, { status: 403 });
  }

  const { nama } = await req.json();
  const validation = validateNama(nama);
  if (!validation.ok) {
    return Response.json({ error: validation.error }, { status: 400 });
  }

  // Case-insensitive check: the DB's unique index is case-sensitive, so
  // "Retail" and "retail" would both be accepted and read as duplicates.
  const { rows: existing } = await query<{ id: string }>(
    `select id from classifications where lower(nama) = lower($1) limit 1`,
    [validation.value]
  );
  if (existing[0]) {
    return Response.json({ error: NAMA_TAKEN }, { status: 409 });
  }

  const { rows } = await query<{ id: string }>(
    `insert into classifications (nama) values ($1) returning id`,
    [validation.value]
  );
  return Response.json({ id: rows[0].id }, { status: 201 });
}
