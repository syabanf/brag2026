import { NextRequest } from "next/server";
import { requireUser } from "@/lib/auth";
import { query } from "@/lib/db";
import { parseBoosterScoring } from "@/lib/booster-input";

export async function PATCH(req: NextRequest, { params }: { params: Promise<{ id: string }> }) {
  const { user } = await requireUser();
  if (user.role !== "admin" && user.role !== "super_admin") {
    return Response.json({ error: "Forbidden" }, { status: 403 });
  }

  const { id } = await params;
  const body = await req.json();
  const { judul, deskripsi, tanggal_mulai, tanggal_berakhir, poin, status, multiplier, band_min, band_max } = body;

  if (status !== undefined) {
    // Toggle status only
    await query(
      `update booster_events set status = $1 where id = $2`,
      [status, id]
    );
    return Response.json({ ok: true });
  }

  if (!judul || !tanggal_mulai || !tanggal_berakhir || poin === undefined) {
    return Response.json({ error: "Judul, tanggal, dan poin wajib diisi." }, { status: 400 });
  }

  const parsed = parseBoosterScoring({ multiplier, band_min, band_max });
  if ("error" in parsed) return Response.json({ error: parsed.error }, { status: 400 });

  await query(`
    update booster_events
    set judul = $1, deskripsi = $2, tanggal_mulai = $3, tanggal_berakhir = $4, poin = $5,
        multiplier = $6, band_min = $7, band_max = $8
    where id = $9
  `, [
    judul.trim(), deskripsi?.trim() ?? null, tanggal_mulai, tanggal_berakhir, Number(poin),
    parsed.multiplier, parsed.band_min, parsed.band_max, id,
  ]);

  return Response.json({ ok: true });
}

export async function DELETE(_req: NextRequest, { params }: { params: Promise<{ id: string }> }) {
  const { user } = await requireUser();
  if (user.role !== "admin" && user.role !== "super_admin") {
    return Response.json({ error: "Forbidden" }, { status: 403 });
  }

  const { id } = await params;
  await query(`delete from booster_events where id = $1`, [id]);
  return Response.json({ ok: true });
}
