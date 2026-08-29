-- Demo facts: accounts, teams, the event calendar, the prize pool. Everything
-- with a scoring rule attached — transactions, visitors, the ledger, badges,
-- raffle tickets — is deliberately absent, because `cmd/seed` produces it by
-- calling the same use cases the running app calls. Rules live in Go; this
-- file only supplies the world they act on.
--
--   bash scripts/seed-demo.sh
--
-- Safe to re-run: it clears everything it creates first. Deterministic on
-- purpose, so the same demo looks identical every time.

do $$
declare
  v_season uuid;
  v_admin  uuid;
  v_klas   uuid[];
begin
  select id into v_season from event_seasons where nama = 'BRAG 2026';
  select id into v_admin  from app_users where email = 'ilham@wit.id';

  if v_season is null then
    raise exception 'Season "BRAG 2026" not found — apply migrations first.';
  end if;

  -- ── Reset ────────────────────────────────────────────────────────────
  -- Ordered so no delete trips a foreign key. This clears the derived rows
  -- too: cmd/seed rebuilds them straight after.
  delete from raffle_tickets  where season_id = v_season;
  delete from prize_pool      where season_id = v_season;
  delete from member_badges;
  delete from score_ledger    where season_id = v_season;
  delete from tyfcb_entries   where season_id = v_season;
  delete from visitors        where season_id = v_season;
  delete from one_to_one_logs where season_id = v_season;
  delete from contact_sphere_members
    where sphere_id in (select id from contact_spheres where season_id = v_season);
  delete from contact_spheres where season_id = v_season;
  delete from weekly_events   where season_id = v_season;
  delete from booster_events  where season_id = v_season;

  -- The season started a week ago, so the app reads as week 2 of 12.
  update event_seasons
  set starts_on = current_date - 7,
      ends_on   = current_date - 7 + interval '12 weeks',
      status    = 'active'
  where id = v_season;

  -- ── Demo accounts ────────────────────────────────────────────────────
  -- One per role, sharing an obvious password. The login screen offers these
  -- as one-click sign-ins, so the addresses are part of the contract: see
  -- frontend/src/lib/demo-accounts.ts.
  insert into app_users (email, password_hash, full_name, role) values
    ('demo.admin@brag2026.id',   crypt('demo123', gen_salt('bf')), 'Demo Admin',   'admin'),
    ('demo.captain@brag2026.id', crypt('demo123', gen_salt('bf')), 'Demo Captain', 'captain'),
    ('demo.member@brag2026.id',  crypt('demo123', gen_salt('bf')), 'Demo Member',  'member')
  on conflict (email) do update set password_hash = excluded.password_hash,
                                    full_name     = excluded.full_name,
                                    role          = excluded.role;

  -- Each persona needs a competition profile or their dashboard is empty.
  insert into members (user_id, season_id, team_id, color_status, is_active)
  select u.id, v_season,
         (select id from teams where season_id = v_season and nama_tim = t.nama),
         t.warna::color_status, true
  from (values
    ('demo.admin@brag2026.id',   'Tim 1', 'hijau'),
    ('demo.captain@brag2026.id', 'Tim 1', 'hijau'),
    ('demo.member@brag2026.id',  'Tim 2', 'kuning')
  ) as t(email, nama, warna)
  join app_users u on u.email = t.email
  on conflict (user_id, season_id) do update set team_id      = excluded.team_id,
                                                 color_status = excluded.color_status;

  -- The seeded superadmin ships without a profile, which leaves their own
  -- dashboard blank; give them one too.
  insert into members (user_id, season_id, team_id, color_status, is_active)
  select v_admin, v_season,
         (select id from teams where season_id = v_season and nama_tim = 'Tim 1'),
         'hijau', true
  on conflict (user_id, season_id) do nothing;

  -- ── Roster shape ─────────────────────────────────────────────────────
  -- Spread classifications and colour status so the leaderboards, the
  -- CAT_CAROUSEL event and the UNDERDOG event all have something to act on.
  -- hashtext keeps it deterministic without depending on row order.
  select array_agg(id order by nama) into v_klas from classifications;

  update members m
  set klasifikasi_id = v_klas[1 + (abs(hashtext(m.id::text)) % array_length(v_klas, 1))],
      color_status = coalesce(
        nullif(m.color_status::text, 'merah')::color_status,
        (array['merah','kuning','hijau','hijau','kuning'])[1 + (abs(hashtext(m.id::text)) % 5)]::color_status)
  where m.season_id = v_season;

  -- Member No.1 of each team leads it.
  update app_users set role = 'captain'
  where email in (select 'm' || t || '1@brag2026.id' from generate_series(1, 10) t);

  -- ── The 12-week event calendar ───────────────────────────────────────
  -- Week 2 — the running week — is FOUNDER, so the multiplier is visible on
  -- the dashboard the moment the demo opens.
  insert into weekly_events (season_id, minggu_ke, event_code, tanggal_mulai, tanggal_selesai)
  select v_season, w,
         (array['SPREAD_LOVE','FOUNDER','VISITOR_BLITZ','CAT_CAROUSEL','UNDERDOG','HIGH_ROLLER',
                'ONE_TO_ONE','STREAK','POWER_TEAM','CLOSING_WEEK','NEW_BLOOD','DOUBLE_UP'])[w],
         current_date - 7 + ((w - 1) * 7),
         current_date - 7 + ((w - 1) * 7) + 6
  from generate_series(1, 12) w;

  -- ── Booster announcements ────────────────────────────────────────────
  -- Three running and one finished, so the screen shows both states.
  insert into booster_events (season_id, judul, deskripsi, tanggal_mulai, tanggal_berakhir, poin, status) values
    (v_season, 'Founder''s Frenzy', 'Semua skor minggu ini dikali 1,5.',                 current_date - 2,  current_date + 4, 50, 'aktif'),
    (v_season, 'Visitor Rush',      'Bawa tamu baru, poin visitor dikali 1,5.',          current_date - 1,  current_date + 5, 30, 'aktif'),
    (v_season, 'High Roller Day',   'TYFCB tunggal tertinggi hari itu dapat +50 flat.',  current_date,      current_date + 7, 50, 'aktif'),
    (v_season, 'One-to-One Sprint', 'Catat 1-on-1 dengan sesama member.',                current_date - 14, current_date - 8, 30, 'nonaktif');

  -- ── Contact spheres ──────────────────────────────────────────────────
  -- POWER_TEAM week pays 1.5× when both sides of a transaction sit in the
  -- same sphere, so the demo needs a few that actually overlap.
  insert into contact_spheres (season_id, nama, deskripsi) values
    (v_season, 'Properti & Interior', 'Developer, kontraktor, interior, furnitur.'),
    (v_season, 'Acara & Hospitality', 'Katering, dekorasi, fotografi, venue.'),
    (v_season, 'Bisnis & Keuangan',   'Konsultan, pajak, notaris, asuransi.');

  -- Three classifications per sphere, taken in catalogue order so a re-run
  -- pairs the same trades together. Spheres are numbered by name for the
  -- same reason: insertion order is not a guarantee.
  insert into contact_sphere_members (sphere_id, klasifikasi_id)
  select ranked.id, v_klas[n]
  from (
    select id, row_number() over (order by nama) as rank
    from contact_spheres where season_id = v_season
  ) ranked
  join lateral generate_series(
    1 + ((ranked.rank::int - 1) * 3),
    least(array_length(v_klas, 1), ranked.rank::int * 3)
  ) as n on true;

  -- ── Prize pool ───────────────────────────────────────────────────────
  insert into prize_pool (season_id, nama_hadiah, deskripsi, nilai_estimasi, alokasi, status) values
    (v_season, 'Grand Prize: Paket Umrah', 'Untuk peraih Team Overall #1.', 35000000, 'kategori', 'approved'),
    (v_season, 'MacBook Air M3',           'Individu Overall #1.',          22000000, 'kategori', 'approved'),
    (v_season, 'iPad Pro 11"',             'Individu TYFCB #1.',            18000000, 'kategori', 'approved'),
    (v_season, 'Sepeda Lipat Brompton',    'Individu Visitor #1.',          30000000, 'kategori', 'approved'),
    (v_season, 'Voucher Belanja 5 Juta',   'Diundi untuk semua member.',     5000000, 'undian',   'approved'),
    (v_season, 'Smartwatch Series 10',     'Diundi untuk semua member.',     8000000, 'undian',   'approved'),
    (v_season, 'Weekend Staycation',       'Diundi untuk semua member.',     6500000, 'undian',   'approved');

  -- Member donations: one approved (which earns PATRON), one still waiting.
  insert into prize_pool (season_id, nama_hadiah, deskripsi, nilai_estimasi, donatur_id, alokasi, status)
  select v_season, d.nama, d.deskripsi, d.nilai, m.id,
         d.alokasi::prize_alokasi, d.status::prize_status
  from (values
    ('Voucher Konsultasi Bisnis', 'Sesi 2 jam bersama donatur.',    2500000::numeric, 'undian',   'approved', 'm11@brag2026.id'),
    ('Paket Catering 50 Porsi',   'Menunggu persetujuan panitia.',  4000000::numeric, 'kategori', 'pending',  'm21@brag2026.id')
  ) as d(nama, deskripsi, nilai, alokasi, status, email)
  join app_users u on u.email = d.email
  join members m on m.user_id = u.id and m.season_id = v_season;

  raise notice 'Demo facts seeded. Run cmd/seed to generate activity.';
end;
$$;
