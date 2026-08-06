-- Pair penalty toggle, per season.
-- When enabled (default), repeat TYFCB between the same buyer→seller pair decays
-- per spec §4.1: ordinal 1-2 = 1.0x, 3-5 = 0.7x, 6+ = 0.5x.
-- When disabled, every entry scores at full band.

alter table event_seasons
  add column if not exists pair_penalty_enabled boolean not null default true;
