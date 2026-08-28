#!/usr/bin/env bash
# Loads sample data so every screen has real numbers. Opt-in and re-runnable —
# it clears its own rows first. Never run this against production.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
seed="$root/backend/seeds/demo.sql"

if docker compose ps --status running --services 2>/dev/null | grep -qx db; then
  echo "→ seeding the compose database"
  docker compose exec -T db psql -v ON_ERROR_STOP=1 -U brag -d brag_dev < "$seed"
else
  : "${DATABASE_URL:?compose is not running; set DATABASE_URL to seed directly}"
  echo "→ seeding \$DATABASE_URL"
  psql -v ON_ERROR_STOP=1 "$DATABASE_URL" -f "$seed"
fi

cat <<'ACCOUNTS'

Demo accounts (password: demo123)
  demo.admin@brag2026.id     admin    — verification, master data, passes
  demo.captain@brag2026.id   captain  — files on behalf of Tim 1
  demo.member@brag2026.id    member   — ordinary member on Tim 2

Also seeded: ilham@wit.id / admin123, and m<team><member>@brag2026.id / member123
ACCOUNTS
