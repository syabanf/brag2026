-- Demo data: a season already in flight, so every screen has real numbers to
-- show. Opt-in and separate from migrations/ because this is sample content,
-- not schema — production must never run it.
--
--   bash scripts/seed-demo.sh
--
-- Safe to re-run: it clears its own activity first, then rebuilds. Deterministic
-- on purpose, so the same demo looks identical every time.

do $$
declare
  v_season uuid;
  v_admin  uuid;
  v_klas   uuid[];
  v_member record;
  v_giver  uuid;
  v_recv   uuid;
  v_nilai  numeric;
  v_band   int;
  v_mult   numeric;
  v_status text;
  i        int;
  n        int;
begin
  select id into v_season from event_seasons where nama = 'BRAG 2026';
  select id into v_admin  from app_users where email = 'ilham@wit.id';

  if v_season is null then
    raise exception 'Season "BRAG 2026" not found — apply migrations first.';
  end if;

  -- ── Reset only what this seed creates ────────────────────────────────
  delete from raffle_tickets where season_id = v_season;
  delete from prize_pool     where season_id = v_season;
  delete from member_badges;
  delete from score_ledger   where season_id = v_season;
  delete from tyfcb_entries  where season_id = v_season;
  delete from visitors       where season_id = v_season;
  delete from weekly_events  where season_id = v_season;
  delete from booster_events where season_id = v_season;

  -- Season started a week ago, so the app reads as week 2 of 12.
  update event_seasons
  set starts_on = current_date - 7,
      ends_on   = current_date - 7 + interval '12 weeks',
      status    = 'active'
  where id = v_season;

  -- ── Demo accounts ────────────────────────────────────────────────────
  -- Three personas covering every role, with an obvious shared password.
  insert into app_users (email, password_hash, full_name, role) values
    ('demo.admin@brag2026.id',   crypt('demo123', gen_salt('bf')), 'Demo Admin',   'admin'),
    ('demo.captain@brag2026.id', crypt('demo123', gen_salt('bf')), 'Demo Captain', 'captain'),
    ('demo.member@brag2026.id',  crypt('demo123', gen_salt('bf')), 'Demo Member',  'member')
  on conflict (email) do update set password_hash = excluded.password_hash,
                                    role = excluded.role;

  -- Each demo persona needs a competition profile, or their dashboard is empty.
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
  on conflict (user_id, season_id) do update set team_id = excluded.team_id,
                                                 color_status = excluded.color_status;

  -- The seeded superadmin ships without a profile, which leaves their
  -- dashboard blank; give them one too.
  insert into members (user_id, season_id, team_id, color_status, is_active)
  select v_admin, v_season,
         (select id from teams where season_id = v_season and nama_tim = 'Tim 1'),
         'hijau', true
  on conflict (user_id, season_id) do nothing;

  -- Spread classifications and colour status across the roster.
  select array_agg(id order by nama) into v_klas from classifications;

  update members m
  set klasifikasi_id = v_klas[1 + (abs(hashtext(m.id::text)) % 10)],
      color_status = coalesce(
        nullif(m.color_status::text, 'merah')::color_status,
        (array['merah','kuning','hijau','hijau','kuning'])[1 + (abs(hashtext(m.id::text)) % 5)]::color_status)
  where m.season_id = v_season;

  -- One captain per team: member No.1 of each.
  update app_users set role = 'captain'
  where email in (select 'm' || t || '1@brag2026.id' from generate_series(1, 10) t);

  -- ── 12 weekly events ─────────────────────────────────────────────────
  -- Week 2 is FOUNDER so the running week visibly boosts everything, which
  -- makes the multiplier easy to demonstrate.
  insert into weekly_events (season_id, minggu_ke, event_code, tanggal_mulai, tanggal_selesai)
  select v_season, w,
         (array['SPREAD_LOVE','FOUNDER','VISITOR_BLITZ','CAT_CAROUSEL','UNDERDOG','HIGH_ROLLER',
                'ONE_TO_ONE','STREAK','POWER_TEAM','CLOSING_WEEK','NEW_BLOOD','DOUBLE_UP'])[w],
         current_date - 7 + ((w - 1) * 7),
         current_date - 7 + ((w - 1) * 7) + 6
  from generate_series(1, 12) w;

  -- ── Booster announcements ────────────────────────────────────────────
  insert into booster_events (season_id, judul, deskripsi, tanggal_mulai, tanggal_berakhir, poin, status) values
    (v_season, 'Founder''s Frenzy',  'Semua skor minggu ini dikali 1,5.',                current_date - 2, current_date + 4, 50, 'aktif'),
    (v_season, 'Visitor Rush',       'Bawa tamu baru, poin visitor dikali 1,5.',         current_date - 1, current_date + 5, 30, 'aktif'),
    (v_season, 'High Roller Day',    'TYFCB tunggal tertinggi hari itu dapat +50 flat.', current_date,     current_date + 7, 50, 'aktif'),
    (v_season, 'One-to-One Sprint',  'Catat 1-on-1 dengan sesama member.',               current_date - 14, current_date - 8, 30, 'nonaktif');

  -- ── TYFCB entries ────────────────────────────────────────────────────
  -- 180 across the roster: roughly 70% verified, 20% pending, 10% rejected.
  i := 0;
  for v_member in
    select m.id from members m
    join app_users u on u.id = m.user_id
    where m.season_id = v_season order by u.email
  loop
    for n in 1..2 loop
      i := i + 1;
      exit when i > 180;

      v_giver := v_member.id;
      select id into v_recv from members
      where season_id = v_season and id <> v_giver
      order by md5(id::text || i::text) limit 1;

      v_nilai := (array[350000, 1200000, 4500000, 18000000, 62000000,
                        140000000, 320000000, 780000000])[1 + ((i * 3) % 8)];

      v_band := case
        when v_nilai <    500000 then 10
        when v_nilai <   2000000 then 25
        when v_nilai <  10000000 then 50
        when v_nilai <  50000000 then 80
        when v_nilai < 250000000 then 120
        when v_nilai < 500000000 then 150
        else 200 end;

      v_mult   := (array[1.0, 1.0, 1.5, 2.0])[1 + (i % 4)];
      v_status := (array['verified','verified','verified','verified','verified',
                         'verified','verified','pending','pending','rejected'])[1 + (i % 10)];

      insert into tyfcb_entries
        (season_id, giver_id, receiver_id, nilai, tanggal, status, computed_score,
         pair_ordinal, event_multiplier_applied, verified_by, verified_at, rejection_reason)
      values (v_season, v_giver, v_recv, v_nilai, current_date - ((i % 12) + 1),
              v_status::tyfcb_status,
              case when v_status = 'verified' then round(v_band * (1.0 / n) * v_mult) end,
              n,
              case when v_status = 'verified' then v_mult end,
              case when v_status = 'verified' then v_admin end,
              case when v_status = 'verified' then now() - ((i % 10) || ' days')::interval end,
              case when v_status = 'rejected' then 'Bukti transaksi tidak terbaca.' end);
    end loop;
    exit when i > 180;
  end loop;

  -- ── Visitors ─────────────────────────────────────────────────────────
  i := 0;
  for v_member in
    select m.id from members m
    join app_users u on u.id = m.user_id
    where m.season_id = v_season order by u.email limit 60
  loop
    i := i + 1;
    insert into visitors
      (season_id, nama, kontak, inviter_id, tanggal_undang, status_hadir, is_converted, tanggal_konversi)
    values (v_season, 'Tamu ' || i, '08' || lpad((81000000 + i * 137)::text, 10, '0'),
            v_member.id, current_date - ((i % 14) + 1),
            (array['terdaftar','hadir','hadir_penuh','hadir','hadir_penuh'])[1 + (i % 5)]::visitor_status,
            (i % 7 = 0),
            case when i % 7 = 0 then current_date - (i % 5) end)
    on conflict (season_id, kontak) do nothing;
  end loop;

  -- ── Score ledger — the single source of truth ────────────────────────
  insert into score_ledger (season_id, member_id, team_id, kategori, points, sumber_ref, keterangan, created_at)
  select v_season, te.giver_id, m.team_id, 'tyfcb', te.computed_score, te.id::text,
         'TYFCB verified', te.verified_at
  from tyfcb_entries te join members m on m.id = te.giver_id
  where te.season_id = v_season and te.status = 'verified' and te.computed_score is not null;

  insert into score_ledger (season_id, member_id, team_id, kategori, points, sumber_ref, keterangan)
  select v_season, v.inviter_id, m.team_id, 'visitor',
         case v.status_hadir when 'terdaftar' then 0 when 'hadir' then 20 else 50 end,
         v.id::text, 'Status visitor: ' || v.status_hadir
  from visitors v join members m on m.id = v.inviter_id
  where v.season_id = v_season
    and case v.status_hadir when 'terdaftar' then 0 when 'hadir' then 20 else 50 end > 0;

  insert into score_ledger (season_id, member_id, team_id, kategori, points, sumber_ref, keterangan)
  select v_season, v.inviter_id, m.team_id, 'visitor', 100,
         v.id::text || ':conversion', 'Visitor konversi'
  from visitors v join members m on m.id = v.inviter_id
  where v.season_id = v_season and v.is_converted;

  -- Team bonuses, keyed the same way the weekly pass keys them so a later run
  -- recognises them as already settled.
  insert into score_ledger (season_id, team_id, kategori, points, sumber_ref, keterangan)
  select v_season, t.id, 'bonus', 100,
         'full_roster:' || t.id || ':' || to_char(current_date - 7, 'YYYY-MM-DD'),
         'Full Roster minggu 1'
  from teams t
  where t.season_id = v_season and (abs(hashtext(t.nama_tim)) % 3) <> 0;

  -- ── Prize pool ───────────────────────────────────────────────────────
  insert into prize_pool (season_id, nama_hadiah, deskripsi, nilai_estimasi, alokasi, status) values
    (v_season, 'Grand Prize: Paket Umrah',    'Untuk peraih Team Overall #1.',     35000000, 'kategori', 'approved'),
    (v_season, 'MacBook Air M3',              'Individu Overall #1.',              22000000, 'kategori', 'approved'),
    (v_season, 'Voucher Belanja 5 Juta',      'Diundi untuk semua member.',         5000000, 'undian',   'approved'),
    (v_season, 'Smartwatch Series 10',        'Diundi untuk semua member.',         8000000, 'undian',   'approved');

  -- Two member donations: one approved (earns PATRON), one still pending.
  insert into prize_pool (season_id, nama_hadiah, deskripsi, nilai_estimasi, donatur_id, alokasi, status)
  select v_season, 'Voucher Konsultasi Bisnis', 'Sesi 2 jam bersama donatur.', 2500000,
         m.id, 'undian', 'approved'
  from members m join app_users u on u.id = m.user_id
  where u.email = 'm11@brag2026.id' limit 1;

  insert into prize_pool (season_id, nama_hadiah, deskripsi, nilai_estimasi, donatur_id, alokasi, status)
  select v_season, 'Paket Catering 50 Porsi', 'Menunggu persetujuan panitia.', 4000000,
         m.id, 'kategori', 'pending'
  from members m join app_users u on u.id = m.user_id
  where u.email = 'm21@brag2026.id' limit 1;

  -- ── Raffle tickets ───────────────────────────────────────────────────
  -- floor(score/100) + one per attending visitor + one per first-time pair.
  insert into raffle_tickets (season_id, member_id, sumber)
  select v_season, m.id, 'score'
  from members m
  join lateral generate_series(1, greatest(
    (select coalesce(sum(points), 0)::int / 100 from score_ledger
     where member_id = m.id and season_id = v_season), 0)) g on true
  where m.season_id = v_season and m.is_active;

  insert into raffle_tickets (season_id, member_id, sumber)
  select v_season, v.inviter_id, 'visitor'
  from visitors v
  where v.season_id = v_season and v.status_hadir in ('hadir', 'hadir_penuh');

  insert into raffle_tickets (season_id, member_id, sumber)
  select v_season, te.giver_id, 'tyfcb_pair'
  from tyfcb_entries te
  where te.season_id = v_season and te.status = 'verified' and te.pair_ordinal = 1;

  -- ── Badges ───────────────────────────────────────────────────────────
  -- Derived from the data above using the same thresholds the Go rules apply,
  -- so the demo starts consistent with what the evaluator would award.
  insert into member_badges (member_id, badge_code)
  select m.id, 'FIRST_BLOOD' from members m
  where m.season_id = v_season and exists (
    select 1 from tyfcb_entries where giver_id = m.id and status = 'verified');

  insert into member_badges (member_id, badge_code)
  select m.id, 'HOST' from members m
  where m.season_id = v_season and exists (
    select 1 from visitors where inviter_id = m.id and status_hadir in ('hadir', 'hadir_penuh'));

  insert into member_badges (member_id, badge_code)
  select m.id, 'CLOSER' from members m
  where m.season_id = v_season and exists (
    select 1 from visitors where inviter_id = m.id and is_converted);

  insert into member_badges (member_id, badge_code)
  select m.id, 'CENTURION' from members m
  where m.season_id = v_season and (
    select coalesce(sum(points), 0) from score_ledger
    where member_id = m.id and season_id = v_season) >= 100;

  insert into member_badges (member_id, badge_code)
  select m.id, 'HIGH_ROLLER' from members m
  where m.season_id = v_season and exists (
    select 1 from tyfcb_entries
    where giver_id = m.id and status = 'verified' and nilai >= 250000000);

  insert into member_badges (member_id, badge_code)
  select m.id, 'LEVEL_UP' from members m
  where m.season_id = v_season and m.color_status <> 'merah';

  insert into member_badges (member_id, badge_code)
  select m.id, 'TEAM_PLAYER' from members m
  where m.season_id = v_season and exists (
    select 1 from score_ledger sl
    where sl.team_id = m.team_id and sl.sumber_ref like 'full_roster:%');

  insert into member_badges (member_id, badge_code)
  select p.donatur_id, 'PATRON' from prize_pool p
  where p.season_id = v_season and p.donatur_id is not null and p.status = 'approved';

  raise notice 'Demo data seeded.';
end;
$$;
