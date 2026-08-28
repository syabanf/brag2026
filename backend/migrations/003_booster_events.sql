-- Migration: booster_events
-- Apply manually: psql $DATABASE_URL -f db/local/003_booster_events.sql

create table if not exists booster_events (
  id               uuid        primary key default gen_random_uuid(),
  season_id        uuid        not null references event_seasons(id),
  judul            text        not null,
  deskripsi        text,
  tanggal_mulai    date        not null,
  tanggal_berakhir date        not null,
  poin             integer     not null default 0,
  status           text        not null default 'aktif'
                               check (status in ('aktif', 'nonaktif')),
  created_at       timestamptz not null default now(),
  check (tanggal_berakhir >= tanggal_mulai)
);
