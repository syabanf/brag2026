-- API keys for machine access.
--
-- A key acts as an existing user rather than carrying its own permissions.
-- That way there is exactly one authorization model in the codebase: the role
-- checks that already guard every route apply unchanged, and a key can never
-- reach something its owner could not.
--
-- read_only is the one thing a key adds on top. Most integrations only read,
-- and a leaked read-only key cannot award points — which, on an append-only
-- ledger, would otherwise be permanent.

create table api_keys (
  id       uuid primary key default gen_random_uuid(),
  nama     text not null,
  -- The first characters of the key, stored in clear so a person can tell
  -- their keys apart in a list. Not a secret: it identifies, it does not open.
  prefix   text not null unique,
  -- SHA-256 of the whole key. Keys are high-entropy, so a fast digest is the
  -- right choice here — bcrypt exists to slow down guessing at human-chosen
  -- passwords, and would only tax every authenticated request.
  key_hash text not null unique,

  -- The key authenticates as this user and inherits their role.
  user_id   uuid not null references app_users(id) on delete cascade,
  read_only boolean not null default true,

  last_used_at timestamptz,
  expires_at   timestamptz,
  revoked_at   timestamptz,
  revoked_by   uuid references app_users(id),

  created_by uuid references app_users(id),
  created_at timestamptz not null default now()
);

-- Every authenticated request looks a key up by its digest.
create index api_keys_hash_idx on api_keys (key_hash);
-- The management screen lists a user's keys, newest first.
create index api_keys_user_idx on api_keys (user_id, created_at desc);
