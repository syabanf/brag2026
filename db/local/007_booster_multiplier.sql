-- Booster multiplier, per band.
-- Turns booster_events from an announcement board into a scoring input: a TYFCB entry
-- dated inside an active booster whose band range covers the transaction value is
-- multiplied by that booster's multiplier (spec §4.1 factor M).
--
-- Band range follows TYFCB_BANDS semantics: band_min inclusive, band_max exclusive.
-- Both NULL means the booster applies to every band.
-- Overlapping boosters resolve to the highest multiplier, never the product.
--
-- Apply manually: psql $DATABASE_URL -f db/local/007_booster_multiplier.sql

alter table booster_events
  add column if not exists multiplier numeric(4,2) not null default 1.0,
  add column if not exists band_min   bigint,
  add column if not exists band_max   bigint;

do $$
begin
  if not exists (
    select 1 from pg_constraint where conname = 'booster_events_multiplier_check'
  ) then
    alter table booster_events
      add constraint booster_events_multiplier_check
      check (multiplier >= 1.0 and multiplier <= 10.0);
  end if;

  if not exists (
    select 1 from pg_constraint where conname = 'booster_events_band_check'
  ) then
    alter table booster_events
      add constraint booster_events_band_check
      check (band_min is null or band_max is null or band_max > band_min);
  end if;
end $$;

-- Lookup path for the scoring engine: active boosters covering a date.
create index if not exists booster_events_active_range_idx
  on booster_events (season_id, tanggal_mulai, tanggal_berakhir)
  where status = 'aktif';
