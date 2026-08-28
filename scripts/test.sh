#!/usr/bin/env bash
# Tests and production builds for both services.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "── backend tests ────────────────────────────────────"
cd "$root/backend"
go test ./...
go build -o /dev/null ./cmd/api

echo "── frontend build ───────────────────────────────────"
cd "$root/frontend"
npm run --silent build

echo "=== Test gate PASS ==="
