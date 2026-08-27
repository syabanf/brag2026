import { NextRequest } from "next/server";
import { requireUser } from "@/lib/auth";
import { query } from "@/lib/db";
import { parseBoosterScoring } from "@/lib/booster-input";

export async function GET() {
  const { user } = await requireUser();
  if (user.role !== "admin" && user.role !== "super_admin") {
    return Response.json({ error: "Forbidden" }, { status: 403 });
  }

  const { rows } = await query(`
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
      be.status,
      be.created_at
    from booster_events be
    join event_seasons es on es.id = be.season_id
    where es.nama = 'BRAG 2026'
    order by be.tanggal_mulai desc
  `, []);

  return Response.json({ boosters: rows });
}

export async function POST(req: NextRequest) {
  const { user } = await requireUser();
  if (user.role !== "admin" && user.role !== "super_admin") {
    return Response.json({ error: "Forbidden" }, { status: 403 });
  }

  const body = await req.json();
  const { judul, deskripsi, tanggal_mulai, tanggal_berakhir, poin, multiplier, band_min, band_max } = body;

  if (!judul || !tanggal_mulai || !tanggal_berakhir || poin === undefined) {
    return Response.json({ error: "Judul, tanggal, dan poin wajib diisi." }, { status: 400 });
  }
  if (new Date(tanggal_berakhir) < new Date(tanggal_mulai)) {
    return Response.json({ error: "Tanggal berakhir harus setelah tanggal mulai." }, { status: 400 });
  }

  const parsed = parseBoosterScoring({ multiplier, band_min, band_max });
  if ("error" in parsed) return Response.json({ error: parsed.error }, { status: 400 });

  const { rows: season } = await query<{ id: string }>(
    `select id from event_seasons where nama = 'BRAG 2026' limit 1`, []
  );
  if (!season[0]) return Response.json({ error: "Season tidak ditemukan." }, { status: 404 });

  const { rows } = await query<{ id: string }>(`
    insert into booster_events
      (season_id, judul, deskripsi, tanggal_mulai, tanggal_berakhir, poin, multiplier, band_min, band_max)
    values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    returning id
  `, [
    season[0].id, judul.trim(), deskripsi?.trim() ?? null, tanggal_mulai, tanggal_berakhir,
    Number(poin), parsed.multiplier, parsed.band_min, parsed.band_max,
  ]);

  return Response.json({ id: rows[0].id }, { status: 201 });
}
