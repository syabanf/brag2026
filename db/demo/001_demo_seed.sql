-- Demo seed — populates a season mid-flight so every screen has real numbers.
-- Applied ONLY to the in-memory PGlite database used by demo mode.
-- Deterministic on purpose: the same demo must look identical on every boot.

do $$
declare
  v_season   uuid;
  v_admin    uuid;
  v_member   record;
  v_giver    uuid;
  v_receiver uuid;
  v_nilai    numeric;
  v_band     int;
  v_mult     numeric;
  v_score    int;
  v_status   text;
  v_klas     uuid[];
  i          int;
  n          int;
begin
  select id into v_season from event_seasons where nama = 'BRAG 2026';
  select id into v_admin  from app_users     where email = 'ilham@wit.id';

  -- Season started a week ago so the app reads as "week 2 of 12".
  update event_seasons set starts_on = current_date - 7,
                           ends_on   = current_date - 7 + interval '12 weeks'
  where id = v_season;

  select array_agg(id order by nama) into v_klas from classifications;

  -- The superadmin ships without a competition profile, which leaves the
  -- member dashboard empty. Demo needs every persona to have real numbers.
  insert into members (user_id, season_id, team_id, color_status, is_active)
  select v_admin, v_season, (select id from teams where season_id = v_season and nama_tim = 'Tim 1'), 'hijau', true
  on conflict (user_id, season_id) do nothing;

  -- Spread classifications and colour status across the roster.
  update members m
  set klasifikasi_id = v_klas[1 + (abs(hashtext(m.id::text)) % 10)],
      color_status   = (array['merah','kuning','hijau','hijau','kuning'])[1 + (abs(hashtext(m.id::text)) % 5)]::color_status
  where m.season_id = v_season;

  -- One captain per team: member No.1 of each team.
  update app_users set role = 'captain'
  where email in (select 'm' || t || '1@brag2026.id' from generate_series(1,10) t);

  -- ── 12 weekly events ──────────────────────────────────────────────
  insert into weekly_events (season_id, minggu_ke, event_code, tanggal_mulai, tanggal_selesai)
  select v_season, w,
         (array['KICKOFF','DOUBLE_TYFCB','VISITOR_RUSH','CAT_CAROUSEL','PAIR_POWER','HIGH_ROLLER',
                'ONE_TO_ONE','STREAK_WEEK','TEAM_SPRINT','CONVERSION','FINAL_PUSH','GRAND_FINALE'])[w],
         current_date - 7 + ((w - 1) * 7),
         current_date - 7 + ((w - 1) * 7) + 6
  from generate_series(1, 12) w
  on conflict (season_id, minggu_ke) do nothing;

  -- ── Booster events ────────────────────────────────────────────────
  insert into booster_events (season_id, judul, deskripsi, tanggal_mulai, tanggal_berakhir, poin, status) values
    (v_season, 'Double TYFCB Week',  'Semua TYFCB terverifikasi minggu ini dikali 2.', current_date - 2, current_date + 4,  50, 'aktif'),
    (v_season, 'Visitor Rush',       'Bawa visitor baru, dapat bonus 30 poin per kepala.', current_date - 1, current_date + 5, 30, 'aktif'),
    (v_season, 'High Roller Bonus',  'TYFCB di atas 50 juta dapat flat bonus 50 poin.',    current_date,     current_date + 7, 50, 'aktif'),
    (v_season, 'One-to-One Sprint',  'Catat 1-on-1 dengan sesama member.',                 current_date - 14, current_date - 8, 30, 'nonaktif');

  -- ── TYFCB entries ─────────────────────────────────────────────────
  -- 180 entries spread across the roster; ~70% verified, 20% pending, 10% rejected.
  i := 0;
  for v_member in
    select m.id, row_number() over (order by u.email) as rn
    from members m join app_users u on u.id = m.user_id
    where m.season_id = v_season
  loop
    for n in 1..2 loop
      i := i + 1;
      exit when i > 180;

      v_giver := v_member.id;
      select id into v_receiver
      from members
      where season_id = v_season and id <> v_giver
      order by md5(id::text || i::text)
      limit 1;

      v_nilai := (array[350000, 1200000, 4500000, 18000000, 62000000, 140000000, 320000000, 780000000])[1 + ((i * 3) % 8)];

      v_band := case
        when v_nilai <    500000 then 10
        when v_nilai <   2000000 then 25
        when v_nilai <  10000000 then 50
        when v_nilai <  50000000 then 80
        when v_nilai < 250000000 then 120
        when v_nilai < 500000000 then 150
        else 200 end;

      v_mult    := (array[1.0, 1.0, 1.5, 2.0])[1 + (i % 4)];
      v_score   := round(v_band * v_mult);
      v_status  := (array['verified','verified','verified','verified','verified','verified','verified','pending','pending','rejected'])[1 + (i % 10)];

      insert into tyfcb_entries (season_id, giver_id, receiver_id, nilai, tanggal, status,
                                 computed_score, pair_ordinal, event_multiplier_applied,
                                 verified_by, verified_at, rejection_reason)
      values (v_season, v_giver, v_receiver, v_nilai,
              current_date - ((i % 12) + 1), v_status::tyfcb_status,
              case when v_status = 'verified' then v_score end,
              n,
              case when v_status = 'verified' then v_mult end,
              case when v_status = 'verified' then v_admin end,
              case when v_status = 'verified' then now() - ((i % 10) || ' days')::interval end,
              case when v_status = 'rejected' then 'Bukti transaksi tidak terbaca.' end);
    end loop;
    exit when i > 180;
  end loop;

  -- ── Visitors ──────────────────────────────────────────────────────
  i := 0;
  for v_member in
    select m.id, row_number() over (order by u.email) as rn
    from members m join app_users u on u.id = m.user_id
    where m.season_id = v_season limit 60
  loop
    i := i + 1;
    insert into visitors (season_id, nama, kontak, inviter_id, tanggal_undang, status_hadir, is_converted, tanggal_konversi)
    values (v_season,
            'Tamu ' || i,
            '08' || lpad((81000000 + i * 137)::text, 10, '0'),
            v_member.id,
            current_date - ((i % 14) + 1),
            (array['terdaftar','hadir','hadir_penuh','hadir','hadir_penuh'])[1 + (i % 5)]::visitor_status,
            (i % 7 = 0),
            case when i % 7 = 0 then current_date - (i % 5) end)
    on conflict (season_id, kontak) do nothing;
  end loop;

  -- ── Score ledger — the single source of truth for all aggregation ──
  insert into score_ledger (season_id, member_id, team_id, kategori, points, sumber_ref, keterangan, created_at)
  select v_season, te.giver_id, m.team_id, 'tyfcb', te.computed_score, te.id::text,
         'TYFCB terverifikasi', te.verified_at
  from tyfcb_entries te join members m on m.id = te.giver_id
  where te.status = 'verified' and te.computed_score is not null;

  insert into score_ledger (season_id, member_id, team_id, kategori, points, sumber_ref, keterangan)
  select v_season, v.inviter_id, m.team_id, 'visitor',
         case v.status_hadir when 'terdaftar' then 20 when 'hadir' then 50 else 150 end,
         v.id::text, 'Visitor ' || v.status_hadir
  from visitors v join members m on m.id = v.inviter_id;

  -- Team bonuses (member_id null = pure team bonus).
  insert into score_ledger (season_id, member_id, team_id, kategori, points, keterangan)
  select v_season, null, t.id, 'bonus', 100, 'Full Roster minggu 1'
  from teams t where t.season_id = v_season and (abs(hashtext(t.nama_tim)) % 3) <> 0;

  insert into score_ledger (season_id, member_id, team_id, kategori, points, keterangan)
  select v_season, null, t.id, 'bonus', 75, 'Level Up tim'
  from teams t where t.season_id = v_season and (abs(hashtext(t.nama_tim)) % 2) = 0;

  -- ── Badges ────────────────────────────────────────────────────────
  insert into member_badges (member_id, badge_code)
  select m.id, b.badge_code
  from members m
  join lateral (
    select badge_code from badges order by md5(badge_code || m.id::text) limit (abs(hashtext(m.id::text)) % 4)
  ) b on true
  where m.season_id = v_season
  on conflict (member_id, badge_code) do nothing;

  raise notice 'Demo seed applied.';
end;
$$;
