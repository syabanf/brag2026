-- BRAG Competition – Schema v2
-- Aligned with official spec (Spesifikasi Aplikasi v1.0)
-- Source of truth for all score aggregation: score_ledger

create extension if not exists pgcrypto;

-- ─────────────────────────────────────────────
-- ENUMS
-- ─────────────────────────────────────────────

create type app_role         as enum ('member', 'admin');
create type color_status     as enum ('merah', 'kuning', 'hijau');
create type tyfcb_status     as enum ('pending', 'verified', 'rejected');
create type visitor_status   as enum ('terdaftar', 'hadir', 'hadir_penuh');
create type ledger_kategori  as enum ('tyfcb', 'visitor', 'bonus');
create type prize_alokasi    as enum ('kategori', 'undian');
create type prize_status     as enum ('pending', 'approved', 'rejected', 'awarded');
create type raffle_sumber    as enum ('score', 'visitor', 'tyfcb_pair');
create type season_status    as enum ('draft', 'active', 'completed');

-- ─────────────────────────────────────────────
-- AUTH LAYER
-- ─────────────────────────────────────────────

create table app_users (
  id            uuid primary key default gen_random_uuid(),
  email         text not null unique,
  password_hash text not null,
  full_name     text not null,
  role          app_role not null default 'member',
  created_at    timestamptz not null default now(),
  updated_at    timestamptz not null default now()
);

create table user_sessions (
  id          uuid primary key default gen_random_uuid(),
  user_id     uuid not null references app_users(id) on delete cascade,
  token_hash  text not null unique,
  expires_at  timestamptz not null,
  created_at  timestamptz not null default now()
);

create index user_sessions_token_idx   on user_sessions (token_hash);
create index user_sessions_expires_idx on user_sessions (expires_at);

-- ─────────────────────────────────────────────
-- COMPETITION LAYER
-- ─────────────────────────────────────────────

create table event_seasons (
  id         uuid primary key default gen_random_uuid(),
  nama       text not null unique,
  starts_on  date,
  ends_on    date,
  status     season_status not null default 'draft',
  created_at timestamptz not null default now()
);

create table classifications (
  id   uuid primary key default gen_random_uuid(),
  nama text not null unique   -- mis. Retail, Jasa, Properti
);

create table teams (
  id        uuid primary key default gen_random_uuid(),
  season_id uuid not null references event_seasons(id) on delete cascade,
  nama_tim  text not null,
  unique (season_id, nama_tim)
);

-- Members = competition profile, linked 1:1 with app_users per season
create table members (
  id                uuid primary key default gen_random_uuid(),
  user_id           uuid not null references app_users(id) on delete cascade,
  season_id         uuid not null references event_seasons(id) on delete cascade,
  team_id           uuid references teams(id) on delete set null,
  klasifikasi_id    uuid references classifications(id) on delete set null,
  color_status      color_status not null default 'merah',
  is_active         boolean not null default true,
  created_at        timestamptz not null default now(),
  unique (user_id, season_id)
);

-- ─────────────────────────────────────────────
-- TYFCB ENTRIES
-- ─────────────────────────────────────────────

create table tyfcb_entries (
  id                       uuid primary key default gen_random_uuid(),
  season_id                uuid not null references event_seasons(id),
  giver_id                 uuid not null references members(id),   -- penerima score
  receiver_id              uuid not null references members(id),   -- yang dapat closed business
  nilai                    numeric(16, 2) not null check (nilai > 0),
  tanggal                  date not null,
  bukti_url                text,
  status                   tyfcb_status not null default 'pending',
  computed_score           integer,                                 -- diisi sistem saat verified
  pair_ordinal             integer,                                 -- urutan ke-n pasangan giver→receiver
  event_multiplier_applied numeric(4, 2),
  rejection_reason         text,
  verified_by              uuid references app_users(id),
  verified_at              timestamptz,
  created_at               timestamptz not null default now(),
  check (giver_id <> receiver_id)
);

create index tyfcb_entries_season_status_idx on tyfcb_entries (season_id, status);
create index tyfcb_entries_giver_idx         on tyfcb_entries (giver_id, tanggal desc);
create index tyfcb_entries_pair_idx          on tyfcb_entries (giver_id, receiver_id, season_id);

-- ─────────────────────────────────────────────
-- VISITORS
-- ─────────────────────────────────────────────

create table visitors (
  id               uuid primary key default gen_random_uuid(),
  season_id        uuid not null references event_seasons(id),
  nama             text not null,
  kontak           text not null,                                   -- wajib, untuk validasi
  inviter_id       uuid not null references members(id),
  tanggal_undang   date not null,
  status_hadir     visitor_status not null default 'terdaftar',
  is_converted     boolean not null default false,
  tanggal_konversi date,
  created_at       timestamptz not null default now(),
  -- satu kontak hanya bisa didaftarkan satu kali per season
  unique (season_id, kontak)
);

create index visitors_inviter_idx  on visitors (inviter_id);
create index visitors_season_idx   on visitors (season_id, status_hadir);

-- ─────────────────────────────────────────────
-- WEEKLY EVENTS
-- ─────────────────────────────────────────────

-- Daftar 12 event codes (dari spec §4.5):
-- CAT_CAROUSEL, VISITOR_BLITZ, CLOSING_WEEK, SPREAD_LOVE, UNDERDOG,
-- POWER_TEAM, HIGH_ROLLER, NEW_BLOOD, ONE_TO_ONE, DOUBLE_UP, STREAK, FOUNDER

create table weekly_events (
  id                       uuid primary key default gen_random_uuid(),
  season_id                uuid not null references event_seasons(id),
  minggu_ke                integer not null check (minggu_ke between 1 and 12),
  event_code               text not null,
  target_classification_id uuid references classifications(id),    -- hanya untuk CAT_CAROUSEL
  tanggal_mulai            date not null,
  tanggal_selesai          date not null,
  created_at               timestamptz not null default now(),
  check (tanggal_selesai >= tanggal_mulai),
  -- hanya satu event aktif per minggu per season
  unique (season_id, minggu_ke)
);

-- ─────────────────────────────────────────────
-- SCORE LEDGER — satu-satunya sumber kebenaran
-- ─────────────────────────────────────────────

create table score_ledger (
  id          uuid primary key default gen_random_uuid(),
  season_id   uuid not null references event_seasons(id),
  member_id   uuid references members(id),                         -- null jika bonus tim murni
  team_id     uuid references teams(id),
  kategori    ledger_kategori not null,
  points      integer not null,                                    -- bisa negatif untuk koreksi
  sumber_ref  text,                                                -- entry_id / visitor_id / event_id
  keterangan  text,
  created_at  timestamptz not null default now()
);

create index score_ledger_season_member_idx on score_ledger (season_id, member_id);
create index score_ledger_season_team_idx   on score_ledger (season_id, team_id);
create index score_ledger_kategori_idx      on score_ledger (season_id, kategori);

-- ─────────────────────────────────────────────
-- BADGES
-- ─────────────────────────────────────────────

-- Seed: 12 badge codes dari spec §7
create table badges (
  badge_code  text primary key,
  nama        text not null,
  deskripsi   text not null,
  ikon        text                                                  -- emoji atau icon name
);

create table member_badges (
  id          uuid primary key default gen_random_uuid(),
  member_id   uuid not null references members(id) on delete cascade,
  badge_code  text not null references badges(badge_code),
  earned_at   timestamptz not null default now(),
  unique (member_id, badge_code)
);

-- ─────────────────────────────────────────────
-- PRIZE POOL
-- ─────────────────────────────────────────────

create table prize_pool (
  id               uuid primary key default gen_random_uuid(),
  season_id        uuid not null references event_seasons(id),
  nama_hadiah      text not null,
  deskripsi        text,
  nilai_estimasi   numeric(14, 2),
  donatur_id       uuid references members(id),                    -- null jika seed panitia
  alokasi          prize_alokasi not null,
  kategori_target  text,                                           -- jika alokasi=kategori
  status           prize_status not null default 'pending',
  pemenang_id      uuid references members(id),
  created_at       timestamptz not null default now()
);

-- ─────────────────────────────────────────────
-- RAFFLE TICKETS
-- ─────────────────────────────────────────────

create table raffle_tickets (
  id         uuid primary key default gen_random_uuid(),
  season_id  uuid not null references event_seasons(id),
  member_id  uuid not null references members(id),
  sumber     raffle_sumber not null,
  created_at timestamptz not null default now()
);

create index raffle_tickets_member_idx on raffle_tickets (season_id, member_id);

-- ─────────────────────────────────────────────
-- TRIGGERS
-- ─────────────────────────────────────────────

create or replace function set_updated_at()
returns trigger language plpgsql as $$
begin new.updated_at = now(); return new; end; $$;

create trigger app_users_set_updated_at
before update on app_users
for each row execute function set_updated_at();

-- ─────────────────────────────────────────────
-- SEED DATA
-- ─────────────────────────────────────────────

-- Admin account
insert into app_users (email, password_hash, full_name, role) values (
  'ilham@wit.id',
  crypt('admin123', gen_salt('bf')),
  'Ilham',
  'admin'
) on conflict (email) do update
  set password_hash = excluded.password_hash,
      full_name     = excluded.full_name,
      role          = excluded.role;

-- Season aktif
insert into event_seasons (nama, starts_on, ends_on, status) values
  ('BRAG 2026', current_date, current_date + interval '12 weeks', 'active')
on conflict (nama) do nothing;

-- Classifications (contoh umum BNI Grow)
insert into classifications (nama) values
  ('Retail'), ('Jasa'), ('Properti'), ('Kuliner'), ('Teknologi'),
  ('Kesehatan'), ('Pendidikan'), ('Keuangan'), ('Manufaktur'), ('Lainnya')
on conflict (nama) do nothing;

-- 10 Teams untuk BRAG 2026
insert into teams (season_id, nama_tim)
select es.id, t.nama
from event_seasons es
cross join (values
  ('Tim 1'), ('Tim 2'), ('Tim 3'), ('Tim 4'), ('Tim 5'),
  ('Tim 6'), ('Tim 7'), ('Tim 8'), ('Tim 9'), ('Tim 10')
) as t(nama)
where es.nama = 'BRAG 2026'
on conflict (season_id, nama_tim) do nothing;

-- Badge seed (§7)
insert into badges (badge_code, nama, deskripsi, ikon) values
  ('FIRST_BLOOD',   'First Blood',    'TYFCB verified pertama',                            '🩸'),
  ('HOST',          'Tuan Rumah',     'Visitor pertama mencapai status hadir',              '🏠'),
  ('CLOSER',        'Closer',         'Konversi member pertama',                            '🤝'),
  ('CONNECTOR',     'Connector',      'TYFCB ke 5 receiver berbeda',                        '🔗'),
  ('SPREADER',      'Spreader',       'TYFCB ke 10 receiver berbeda',                       '📡'),
  ('CENTURION',     'Centurion',      'Score individu Overall ≥ 100',                       '💯'),
  ('HAT_TRICK',     'Hat-trick',      '3 visitor mencapai hadir_penuh',                     '🎩'),
  ('HIGH_ROLLER',   'High Roller',    'Punya satu TYFCB ≥ Rp 250 juta',                    '💎'),
  ('STREAK_MASTER', 'Streak Master',  'Log aktivitas berscore 3+ hari berbeda dalam seminggu', '🔥'),
  ('TEAM_PLAYER',   'Team Player',    'Ikut menyumbang ke satu Full Roster timnya',         '👥'),
  ('LEVEL_UP',      'Level Up',       'Status warna pribadi naik satu tingkat',             '⬆️'),
  ('PATRON',        'Patron',         'Menyumbang minimal 1 hadiah yang di-approve',        '🎁')
on conflict (badge_code) do nothing;
