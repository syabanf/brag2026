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
# tsc -b, not --noEmit: with project references the plain form reads a
# cached build and silently skips files.
npx tsc -b --force
npm run --silent lint
echo "  typecheck + lint OK"

echo "=== QA gate PASS ==="
