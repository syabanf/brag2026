-- Production scaffolding.
--
-- Migrations already create the accounts, teams, classifications, the season
-- and the badge catalogue. This adds what the engine needs to actually run a
-- season: the twelve-week event calendar, contact spheres for POWER_TEAM
-- week, and a starter prize pool.
--
--   bash scripts/seed-prod.sh
--
-- It is NOT the demo seed and does the opposite of it in one crucial respect:
-- it never deletes anything. There are no fabricated transactions, visitors,
-- ledger rows, badges or raffle tickets here — in production those start
-- empty, because the competition has not happened yet. Every statement is
-- conditional, so running it against a season already in progress adds only
-- what is missing and leaves the rest untouched.

do $$
declare
  v_season uuid;
  v_start  date;
  v_klas   uuid[];
  v_added  int;
begin
  select id, starts_on into v_season, v_start
  from event_seasons where status = 'active'
  order by starts_on desc limit 1;

  if v_season is null then
    raise exception 'No active season — apply migrations first.';
  end if;

  -- ── The twelve-week event calendar ───────────────────────────────────
  -- Without this no multiplier ever applies and every transaction scores at
  -- its base band, which looks like the engine is broken.
  --
  -- Weeks already scheduled are left alone: rewriting the calendar mid-season
  -- would retroactively change what transactions were worth.
  insert into weekly_events (season_id, minggu_ke, event_code, tanggal_mulai, tanggal_selesai)
  select v_season, w.minggu, w.code,
         v_start + ((w.minggu - 1) * 7),
         v_start + ((w.minggu - 1) * 7) + 6
  from (values
    ( 1, 'SPREAD_LOVE'),
    ( 2, 'FOUNDER'),
    ( 3, 'VISITOR_BLITZ'),
    ( 4, 'CAT_CAROUSEL'),
    ( 5, 'UNDERDOG'),
    ( 6, 'HIGH_ROLLER'),
    ( 7, 'ONE_TO_ONE'),
    ( 8, 'STREAK'),
    ( 9, 'POWER_TEAM'),
    (10, 'CLOSING_WEEK'),
    (11, 'NEW_BLOOD'),
    (12, 'DOUBLE_UP')
  ) as w(minggu, code)
  where not exists (
    select 1 from weekly_events e
    where e.season_id = v_season and e.minggu_ke = w.minggu
  );

  get diagnostics v_added = row_count;
  raise notice 'Weekly events added: %', v_added;

  -- ── Contact spheres ──────────────────────────────────────────────────
  -- POWER_TEAM week pays 1.5× when both sides of a transaction sit in the
  -- same sphere. With none defined that week scores like any other, so the
  -- groupings need to exist before week 9 arrives.
  --
  -- These are a starting point drawn from the classification catalogue. The
  -- admin panel is where they get adjusted to how this chapter actually
  -- refers business.
  select array_agg(id order by nama) into v_klas from classifications;

  -- Seeded only into an empty set. Matching on name would resurrect a sphere
  -- the admin had renamed, and the name is precisely what gets edited.
  if array_length(v_klas, 1) >= 6
     and not exists (select 1 from contact_spheres where season_id = v_season) then
    insert into contact_spheres (season_id, nama, deskripsi)
    select v_season, s.nama, s.deskripsi
    from (values
      ('Properti & Interior', 'Developer, kontraktor, interior, furnitur.'),
      ('Acara & Hospitality', 'Katering, dekorasi, fotografi, venue.'),
      ('Bisnis & Keuangan',   'Konsultan, pajak, notaris, asuransi.')
    ) as s(nama, deskripsi)
    where not exists (
      select 1 from contact_spheres c
      where c.season_id = v_season and c.nama = s.nama
    );

    insert into contact_sphere_members (sphere_id, klasifikasi_id)
    select ranked.id, v_klas[n]
    from (
      select id, row_number() over (order by nama) as rank
      from contact_spheres where season_id = v_season
    ) ranked
    join lateral generate_series(
      1 + ((ranked.rank::int - 1) * 3),
      least(array_length(v_klas, 1), ranked.rank::int * 3)
    ) as n on true
    where not exists (
      select 1 from contact_sphere_members m
      where m.sphere_id = ranked.id and m.klasifikasi_id = v_klas[n]
    );
  end if;

  -- ── Starter prize pool ───────────────────────────────────────────────
  -- Placeholders with round numbers, meant to be edited in the admin panel
  -- before anyone sees them. Six category prizes match the six leaderboards;
  -- the three raffle prizes give the ticket system something to draw for.
  -- A starter pool starts the pool; it does not top it up. Once there is a
  -- single prize this block does nothing, because the alternative is that
  -- renaming "Hadiah Undian 1" causes it to reappear on the next run.
  insert into prize_pool (season_id, nama_hadiah, deskripsi, nilai_estimasi, alokasi, kategori_target, status)
  select v_season, p.nama, p.deskripsi, p.nilai, p.alokasi::prize_alokasi, p.target, 'approved'::prize_status
  from (values
    ('Hadiah Team Overall #1',    'Untuk tim peringkat 1 klasemen keseluruhan.', 35000000::numeric, 'kategori', 'team_overall'),
    ('Hadiah Individu Overall #1','Untuk peraih poin tertinggi keseluruhan.',    22000000::numeric, 'kategori', 'individu_overall'),
    ('Hadiah Team TYFCB #1',      'Untuk tim dengan poin TYFCB tertinggi.',      15000000::numeric, 'kategori', 'team_tyfcb'),
    ('Hadiah Individu TYFCB #1',  'Untuk individu dengan poin TYFCB tertinggi.', 12000000::numeric, 'kategori', 'individu_tyfcb'),
    ('Hadiah Team Visitor #1',    'Untuk tim dengan poin visitor tertinggi.',    15000000::numeric, 'kategori', 'team_visitor'),
    ('Hadiah Individu Visitor #1','Untuk individu dengan poin visitor tertinggi.',12000000::numeric,'kategori', 'individu_visitor'),
    ('Hadiah Undian 1',           'Diundi untuk semua pemegang tiket.',           8000000::numeric, 'undian',   null),
    ('Hadiah Undian 2',           'Diundi untuk semua pemegang tiket.',           6000000::numeric, 'undian',   null),
    ('Hadiah Undian 3',           'Diundi untuk semua pemegang tiket.',           4000000::numeric, 'undian',   null)
  ) as p(nama, deskripsi, nilai, alokasi, target)
  where not exists (select 1 from prize_pool pp where pp.season_id = v_season);

  get diagnostics v_added = row_count;
  raise notice 'Prizes added: %', v_added;

  raise notice 'Production scaffolding ready. No activity data was created or removed.';
end;
$$;
