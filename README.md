# BRAG

BRAG is a member gamification platform for the BNI Grow annual challenge.

The program runs for about three months. Members are grouped into teams by the
committee, and each contribution adds points to their team's leaderboard. The
two scoring categories are **TYFCB** (closed business between members) and
**visitors** brought to meetings, with a weekly event that changes the
multiplier and badges awarded automatically.

## Stack

A Go API in clean architecture behind a Vite React SPA, both on PostgreSQL.

```
backend/    Go 1.26 — chi, pgx, bcrypt
frontend/   Vite + React + TypeScript + Tailwind
docs/       PRD, user stories, acceptance criteria
scripts/    qa / test / security gates
```

## Running it

```bash
docker compose up --build
```

Frontend on http://localhost:5173, API on http://localhost:8081. PostgreSQL is
initialised from `backend/migrations` the first time the volume is created.
Ports are configurable with `WEB_PORT`, `API_PORT` and `DB_PORT`.

Seeded accounts: superadmin `ilham@wit.id` / `admin123`, and 100 members as
`m<team><member>@brag2026.id` (e.g. `m11@brag2026.id`) / `member123`.

### Without Docker

```bash
cd backend  && cp .env.example .env && go run ./cmd/api    # :8080
cd frontend && npm install && npm run dev                   # :5173
```

The frontend reads `VITE_API_URL`. Vite inlines it at build time, so in Docker
it is a build argument rather than a runtime variable.

## Backend architecture

Dependencies point inwards — the domain knows nothing about SQL or HTTP:

```
cmd/api/              composition root; the only place that knows every type
internal/domain/      entities, scoring rules, repository interfaces, errors
internal/usecase/     application rules, orchestrating domain + repositories
internal/repository/  PostgreSQL adapters implementing the domain interfaces
internal/delivery/    HTTP handlers, routing, middleware
migrations/           schema
```

## Scoring

The rules live in `internal/domain/` as pure functions, pinned by tests:

```bash
cd backend && go test ./...
```

**TYFCB** — `round(Band × PairPenalty × EventMultiplier)`

- *Band* steps from 10 points below Rp 500k to 200 points at Rp 500M and above.
- *Pair penalty* is `1/n` for the nth transaction between the same two members,
  so repeat business between the same pair is damped.
- *Event multiplier* comes from the weekly event covering the transaction date.

**Visitor** — cumulative rather than incremental: 0 registered, 20 attended,
50 fully attended, plus 100 on conversion. A status change awards the
*difference*, so correcting a mistake reverses exactly what was given.

**`score_ledger` is append-only.** Reversals are new rows with the opposite
sign, which keeps the audit trail intact and makes every score reproducible
from the ledger alone. Ledger writes share a transaction with the status change
that caused them, so a partial failure cannot skew a leaderboard. Reversals
read what was actually credited rather than recomputing, so points awarded
during a boosted week are given back in full once the boost ends.

### Weekly events

One event may be scheduled per week (`weekly_events`). These codes from the
PRD's event bank are applied automatically:

| Code | Effect |
|------|--------|
| `CAT_CAROUSEL` | TYFCB to the chosen classification ×2 |
| `SPREAD_LOVE` | TYFCB to a new pair (ordinal 1) ×2 |
| `UNDERDOG` | TYFCB to a merah/kuning member ×2 |
| `DOUBLE_UP` | Saturday and Sunday ×1.5 |
| `FOUNDER` | Everything that week ×1.5 |
| `VISITOR_BLITZ` | Visitor points ×1.5 |
| `NEW_BLOOD` | Attendance milestones ×2 |
| `CLOSING_WEEK` | Conversion bonus ×2 |

`POWER_TEAM` needs a contact sphere and `ONE_TO_ONE` needs 1-2-1 logs, neither
of which the schema records. `HIGH_ROLLER` and `STREAK` are flat bonuses that
need an end-of-day or end-of-week pass. All four leave the multiplier at 1
rather than guessing.

### Badges

Ten of the twelve badges are awarded automatically after any scoring change:
First Blood, Tuan Rumah, Closer, Connector, Spreader, Centurion, Hat-trick,
High Roller, Streak Master and Level Up. `TEAM_PLAYER` waits on the weekly
Full Roster pass and `PATRON` on the prize pool; neither is implemented, so
they stay unawarded rather than being approximated.

Evaluation re-derives every badge from a single stats query and awards what is
missing. Awarding is idempotent, and it runs after the transaction commits so a
badge write can never roll back the verification that earned it.

## Gates

```bash
bash scripts/qa.sh              # gofmt, go vet, tsc, oxlint
bash scripts/test.sh            # go test + both production builds
bash scripts/security-check.sh  # tracked-secret scan + dependency audit
```

## Documents

- [Product Requirements](./docs/product/PRD.md) — scoring spec and event bank
- [Project Brief](./docs/project-brief.md)
- [Conventions](./docs/CONVENTIONS.md)
