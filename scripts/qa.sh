#!/usr/bin/env bash
# Static checks for both services. Fails on the first problem.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "── backend ──────────────────────────────────────────"
cd "$root/backend"
gofmt -l . | tee /tmp/brag-gofmt
[ ! -s /tmp/brag-gofmt ] || { echo "FAIL: files above need gofmt"; exit 1; }
go vet ./...
echo "  vet + fmt OK"

echo "── frontend ─────────────────────────────────────────"
cd "$root/frontend"
npx tsc --noEmit
npm run --silent lint
echo "  typecheck + lint OK"

echo "=== QA gate PASS ==="
