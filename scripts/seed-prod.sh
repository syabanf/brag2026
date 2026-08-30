#!/usr/bin/env bash
# Production scaffolding: the event calendar, contact spheres, and a starter
# prize pool. Everything else — accounts, teams, classifications, the season —
# already comes from the migrations.
#
# This never deletes anything and creates no activity data. Safe to re-run: it
# adds only what is missing.
#
#   scripts/seed-prod.sh                    # against the compose database
#   DATABASE_URL=... scripts/seed-prod.sh   # against anything else
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
seed="$root/backend/seeds/production.sql"

if [[ -n "${DATABASE_URL:-}" ]]; then
  echo "→ seeding \$DATABASE_URL"
  psql -v ON_ERROR_STOP=1 "$DATABASE_URL" -f "$seed"
elif docker compose -f "$root/docker-compose.yml" ps --status running --services 2>/dev/null | grep -qx db; then
  echo "→ seeding the compose database"
  docker compose -f "$root/docker-compose.yml" exec -T db \
    psql -v ON_ERROR_STOP=1 -U brag -d brag_dev < "$seed"
else
  echo "compose is not running; set DATABASE_URL to seed directly" >&2
  exit 1
fi

cat <<'ACCOUNTS'

Akun bawaan dari migrasi:
  ilham@wit.id                    admin123     admin
  m<tim><no>@brag2026.id          member123    member  (m11 … m1010)

Ganti kata sandi admin setelah masuk pertama kali.
ACCOUNTS
