-- 002_seed_members.sql
-- Rename teams Tim 1–10, seed 100 members (10 per team), default password: member123
-- Apply: psql -d brag_dev -f db/local/002_seed_members.sql

-- ─────────────────────────────────────────────
-- Step 1: Rename teams Alpha-Juliet → Tim 1-10
-- ─────────────────────────────────────────────

update teams set nama_tim = case nama_tim
  when 'Tim Alpha'   then 'Tim 1'
  when 'Tim Bravo'   then 'Tim 2'
  when 'Tim Charlie' then 'Tim 3'
  when 'Tim Delta'   then 'Tim 4'
  when 'Tim Echo'    then 'Tim 5'
  when 'Tim Foxtrot' then 'Tim 6'
  when 'Tim Golf'    then 'Tim 7'
  when 'Tim Hotel'   then 'Tim 8'
  when 'Tim India'   then 'Tim 9'
  when 'Tim Juliet'  then 'Tim 10'
  else nama_tim
end
where season_id = (select id from event_seasons where nama = 'BRAG 2026');

-- ─────────────────────────────────────────────
-- Step 2: Seed 100 members (10 per team)
-- Email pattern : m<team><member>@brag2026.id  (mis. m11@brag2026.id = Tim1 Member1)
-- Default pass  : member123
-- ─────────────────────────────────────────────

do $$
declare
  v_season_id uuid;
  v_team_id   uuid;
  v_user_id   uuid;
  v_email     text;
  v_name      text;
  tn          int;   -- team number 1..10
  mn          int;   -- member number 1..10
begin
  select id into v_season_id
  from event_seasons where nama = 'BRAG 2026';

  for tn in 1..10 loop
    select id into v_team_id
    from teams
    where season_id = v_season_id and nama_tim = 'Tim ' || tn;

    for mn in 1..10 loop
      v_email := 'm' || tn || mn || '@brag2026.id';
      v_name  := 'Anggota Tim ' || tn || ' No.' || mn;

      insert into app_users (email, password_hash, full_name, role)
      values (v_email, crypt('member123', gen_salt('bf')), v_name, 'member')
      on conflict (email) do nothing;

      select id into v_user_id from app_users where email = v_email;

      insert into members (user_id, season_id, team_id, color_status, is_active)
      values (v_user_id, v_season_id, v_team_id, 'merah', true)
      on conflict (user_id, season_id) do nothing;
    end loop;
  end loop;

  raise notice 'Seeded 100 members across 10 teams.';
end;
$$;
