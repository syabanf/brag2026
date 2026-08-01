import { NextRequest, NextResponse } from "next/server";
import { requireUser } from "@/lib/auth";
import { query } from "@/lib/db";

// Cumulative points a visitor has earned once it reaches a given status.
// Any status change awards (or reverses) the difference between the two.
const STATUS_CUMULATIVE: Record<string, number> = {
  terdaftar: 0,
  hadir: 20,
  hadir_penuh: 50,
};
const STATUS_LABEL: Record<string, string> = {
  terdaftar: "Terdaftar",
  hadir: "Hadir",
  hadir_penuh: "Hadir Penuh",
};
const CONVERSION_POINTS = 100;

export async function PATCH(
  req: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { user } = await requireUser();
  if (user.role !== "admin" && user.role !== "super_admin") {
    return NextResponse.json({ error: "Forbidden" }, { status: 403 });
  }

  const { id } = await params;
  const body: { status_hadir?: string; is_converted?: boolean } = await req.json();

  if (body.status_hadir === undefined && body.is_converted === undefined) {
    return NextResponse.json({ error: "Tidak ada perubahan." }, { status: 400 });
  }

  if (body.status_hadir !== undefined && !(body.status_hadir in STATUS_CUMULATIVE)) {
    return NextResponse.json({ error: "Status tidak valid." }, { status: 400 });
  }

  // Get current visitor with inviter's team_id
  const { rows } = await query<{
    id: string;
    season_id: string;
    inviter_id: string;
    team_id: string | null;
    status_hadir: string;
    is_converted: boolean;
    is_void: boolean;
  }>(`
    select v.id, v.season_id, v.inviter_id, m.team_id,
           v.status_hadir::text as status_hadir, v.is_converted, v.is_void
    from visitors v
    join members m on m.id = v.inviter_id
    where v.id = $1
    limit 1
  `, [id]);

  const visitor = rows[0];
  if (!visitor) return NextResponse.json({ error: "Visitor tidak ditemukan." }, { status: 404 });

  if (visitor.is_void) {
    return NextResponse.json({ error: "Visitor sudah di-void dan tidak bisa diubah." }, { status: 409 });
  }

  // Status change in either direction — the point delta follows the new status.
  if (body.status_hadir && body.status_hadir !== visitor.status_hadir) {
    const from = visitor.status_hadir;
    const to = body.status_hadir;

    // Guarded update doubles as optimistic locking: a concurrent request that
    // already moved the status loses here, so points are never awarded twice.
    const updated = await query(
      `update visitors set status_hadir = $1 where id = $2 and status_hadir = $3`,
      [to, id, from]
    );
    if (updated.rowCount !== 1) {
      return NextResponse.json(
        { error: "Status visitor sudah berubah. Muat ulang halaman." },
        { status: 409 }
      );
    }

    const delta = STATUS_CUMULATIVE[to] - STATUS_CUMULATIVE[from];
    if (delta !== 0) {
      await query(`
        insert into score_ledger (season_id, member_id, team_id, kategori, points, sumber_ref, keterangan)
        values ($1, $2, $3, 'visitor', $4, $5, $6)
      `, [
        visitor.season_id, visitor.inviter_id, visitor.team_id,
        delta, visitor.id,
        `Status visitor: ${STATUS_LABEL[from]} → ${STATUS_LABEL[to]}`,
      ]);
    }
  }

  // Conversion bonus — reversible, since it can be flagged by mistake too.
  if (body.is_converted !== undefined && body.is_converted !== visitor.is_converted) {
    const to = body.is_converted;

    const updated = await query(
      `update visitors
       set is_converted = $1,
           tanggal_konversi = case when $1 then current_date else null end
       where id = $2 and is_converted = $3`,
      [to, id, visitor.is_converted]
    );
    if (updated.rowCount !== 1) {
      return NextResponse.json(
        { error: "Status konversi sudah berubah. Muat ulang halaman." },
        { status: 409 }
      );
    }

    await query(`
      insert into score_ledger (season_id, member_id, team_id, kategori, points, sumber_ref, keterangan)
      values ($1, $2, $3, 'visitor', $4, $5, $6)
    `, [
      visitor.season_id, visitor.inviter_id, visitor.team_id,
      to ? CONVERSION_POINTS : -CONVERSION_POINTS, visitor.id,
      to ? "Visitor konversi" : "Pembatalan konversi visitor",
    ]);
  }

  return NextResponse.json({ ok: true });
}
