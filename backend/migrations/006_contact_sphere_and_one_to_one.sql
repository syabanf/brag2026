-- Backs the last two weekly events the schema could not express.
--
-- POWER_TEAM needs to know which business classifications naturally refer to
-- each other; ONE_TO_ONE needs a record of member meetings. Both are additive,
-- so this applies cleanly to a database that already holds a season.

-- ─────────────────────────────────────────────
-- CONTACT SPHERES  (POWER_TEAM)
-- ─────────────────────────────────────────────

-- A sphere is a named set of classifications that feed each other business —
-- a wedding sphere might hold Photography, Catering and Venue.
create table if not exists contact_spheres (
  id         uuid        primary key default gen_random_uuid(),
  season_id  uuid        not null references event_seasons(id) on delete cascade,
  nama       text        not null,
  deskripsi  text,
  created_at timestamptz not null default now(),
  unique (season_id, nama)
);

-- Membership is many-to-many on purpose: one classification can sit in several
-- spheres (a photographer serves weddings and corporate events alike).
create table if not exists contact_sphere_members (
  sphere_id      uuid not null references contact_spheres(id) on delete cascade,
  klasifikasi_id uuid not null references classifications(id) on delete cascade,
  primary key (sphere_id, klasifikasi_id)
);

create index if not exists contact_sphere_members_klas_idx
  on contact_sphere_members (klasifikasi_id);

-- ─────────────────────────────────────────────
-- ONE-TO-ONE LOGS  (ONE_TO_ONE)
-- ─────────────────────────────────────────────

-- A recorded meeting between two members. The pair is stored in a canonical
-- order (lower uuid first) so the unique constraint catches a duplicate no
-- matter which side files it.
create table if not exists one_to_one_logs (
  id           uuid        primary key default gen_random_uuid(),
  season_id    uuid        not null references event_seasons(id) on delete cascade,
  member_a     uuid        not null references members(id) on delete cascade,
  member_b     uuid        not null references members(id) on delete cascade,
  tanggal      date        not null,
  catatan      text,
  submitted_by uuid        references app_users(id),
  created_at   timestamptz not null default now(),
  check (member_a <> member_b),
  check (member_a < member_b),
  unique (season_id, member_a, member_b, tanggal)
);

create index if not exists one_to_one_logs_season_date_idx
  on one_to_one_logs (season_id, tanggal desc);
create index if not exists one_to_one_logs_member_a_idx on one_to_one_logs (member_a);
create index if not exists one_to_one_logs_member_b_idx on one_to_one_logs (member_b);
