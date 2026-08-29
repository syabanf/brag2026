#!/usr/bin/env bash
# Hardening and stress gate.
#
#   scripts/stress.sh            race-detector suite only (no server needed)
#   scripts/stress.sh --load     also drives a running API and reports latency
#
# The race suite belongs in CI. The load run needs a server with seeded demo
# data, so it stays opt-in.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
base="${BASE_URL:-http://localhost:8081}"
workers="${WORKERS:-100}"
duration="${DURATION:-30s}"
budget="${BUDGET:-500ms}"

echo "── concurrency and hardening suite (-race) ──────────"
cd "$root/backend"
# -count=1 defeats the cache: a race that only shows sometimes is exactly the
# kind a cached PASS would hide.
# -count=5 because the concurrency tests depend on an interleaving: the
# verify-versus-void race took a dozen runs to surface the first time, and a
# single pass would have shipped it.
go test -race -count=5 ./...

if [[ "${1:-}" != "--load" ]]; then
  echo
  echo "=== Stress gate PASS (race suite) ==="
  echo "Add --load to also drive $base"
  exit 0
fi

echo
echo "── load: $workers workers for $duration ─────────────"
if ! curl -fsS -o /dev/null --max-time 5 "$base/api/health"; then
  echo "No server at $base — start it with: docker compose up -d" >&2
  exit 1
fi

go run ./cmd/loadtest \
  -base "$base" -workers "$workers" -duration "$duration" -budget "$budget"

echo
echo "=== Stress gate PASS ==="
