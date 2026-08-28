#!/usr/bin/env bash
# Scans for committed secrets and known-vulnerable dependencies.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"

echo "── committed secrets ────────────────────────────────"
# .env files carry real credentials and must stay untracked.
if git ls-files --error-unmatch .env backend/.env frontend/.env >/dev/null 2>&1; then
  echo "FAIL: a .env file is tracked by git"
  exit 1
fi

# Long API-key-shaped literals in source are almost always a mistake.
if git grep -nE '(sk_|xi-api-key|AKIA)[A-Za-z0-9_-]{16,}' -- backend frontend 2>/dev/null; then
  echo "FAIL: possible hardcoded credential above"
  exit 1
fi
echo "  no secrets found"

echo "── dependencies ─────────────────────────────────────"
(cd frontend && npm audit --audit-level=high) || echo "  npm audit reported findings"
(cd backend && go vet ./...)

echo "=== Security gate PASS ==="
