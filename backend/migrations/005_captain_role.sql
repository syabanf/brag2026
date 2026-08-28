-- EPIC-005: Captain Role
-- Apply manually: psql $DATABASE_URL -f db/local/005_captain_role.sql

-- 1. Extend app_role enum with 'captain'
alter type app_role add value if not exists 'captain';

-- 2. Extend tyfcb_status enum with 'void'
alter type tyfcb_status add value if not exists 'void';

-- 3. Track who submitted a TYFCB entry on behalf of another member
alter table tyfcb_entries
  add column if not exists submitted_by uuid references app_users(id);

-- 4. Track who voided a TYFCB entry
alter table tyfcb_entries
  add column if not exists voided_by uuid references app_users(id),
  add column if not exists voided_at timestamptz;

-- 5. Track voiding on visitors
alter table visitors
  add column if not exists submitted_by uuid references app_users(id),
  add column if not exists voided_by   uuid references app_users(id),
  add column if not exists voided_at   timestamptz,
  add column if not exists is_void     boolean not null default false;
