-- Activates the booster the admin created on 2026-08-27 as a real scoring multiplier.
-- Band 2jt–10jt scores 50, so 2x lands it on 100 poin for the 27 Aug – 3 Sep window.
-- Idempotent: re-running sets the same values.
--
-- Apply manually: psql $DATABASE_URL -f db/local/008_seed_booster_2x_tyfcb.sql

update booster_events
set multiplier = 2.0,
    band_min   = 2000000,
    band_max   = 10000000,
    poin       = 0
where judul = 'Booster 2x score for TYFCB 2-10 Juta'
  and tanggal_mulai    = '2026-08-27'
  and tanggal_berakhir = '2026-09-03';
