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

### Demo data

The migrations create the roster but no activity, so the screens start empty.
Load a season already in flight:

```bash
bash scripts/seed-demo.sh
```

That gives 180 TYFCB entries, 60 visitors, a populated ledger, all twelve
weekly events, a prize pool with member donations, raffle tickets and earned
badges. It is re-runnable — it clears its own rows first — and lives in
`backend/seeds/` rather than `migrations/`, because it is sample content and
must never run in production.

Accounts (all seeded):

| Email | Password | Role |
|-------|----------|------|
| `demo.admin@brag2026.id` | `demo123` | admin |
| `demo.captain@brag2026.id` | `demo123` | captain |
| `demo.member@brag2026.id` | `demo123` | member |
| `ilham@wit.id` | `admin123` | admin |
| `m<team><member>@brag2026.id` | `member123` | member |

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

| `POWER_TEAM` | TYFCB inside a shared contact sphere ×1.5 |

All twelve codes are implemented. `HIGH_ROLLER`, `STREAK` and `ONE_TO_ONE` are
flat bonuses rather than multipliers, so they are settled by the passes below
rather than applied at submission.

**Contact spheres** group the classifications that naturally refer each other
business — a wedding sphere might hold Photography, Catering and Venue. Admins
manage them under Event & Bonus; `POWER_TEAM` rewards a transaction whose two
sides share one.

**One-to-one logs** record member meetings. They carry no points by themselves;
during a `ONE_TO_ONE` week, a pair that both met and closed verified business
earns +30 each.

### Periodic bonuses

Some bonuses depend on a whole day or week, so an admin settles them from
**Event & Bonus**:

| Bonus | Trigger | Value |
|-------|---------|-------|
| Full Roster | every active member of a team scored that week | team +100 |
| Naik Level | admin raises a member's colour status | team +75 / +150 |
| Streak Week | member scored on 3+ distinct days | +40 |
| High Roller Day | largest single verified TYFCB of the day | +50 |
| 1-2-1 Payoff | a logged meeting closed business the same week | +30 each |

Each pass is keyed by period, so re-running one is a no-op instead of a second
payment.

### Prize pool

The committee seeds prizes and members donate for approval. Half the pool goes
to leaderboard category winners, the rest is raffled. Tickets are
`floor(score/100)` + one per attending visitor + one per first-time pair;
issuing rewrites entitlements rather than appending, so the pass stays
re-runnable without inflating anyone's odds.

### Badges

All twelve badges are awarded automatically after any scoring change —
`TEAM_PLAYER` when the member's team collects a Full Roster bonus, `PATRON`
when an admin approves their donation to the prize pool.

Evaluation re-derives every badge from a single stats query and awards what is
missing. Awarding is idempotent, and it runs after the transaction commits so a
badge write can never roll back the verification that earned it.

## Screens

Member — dashboard, leaderboard (plus a public, link-shareable one), submit
(TYFCB, visitor, 1-2-1), booster, awards, prize pool, activity feed, history,
profile. A notification bell in the header surfaces recent activity.
Captain — files TYFCB and visitors for their team, resets team passwords.
Admin — TYFCB verification, visitors, members, teams, classifications,
boosters, weekly events, contact spheres, scoring passes and the prize pool.

A guided tour sits behind the compass button in the header. Its nine steps and
their narration come from the API, so caption and voice cannot drift.

## Listing endpoints

Admin lists are paged and filtered server-side. `?limit=` and `?offset=` (or
`?page=`) window the result; the response carries `total` so a screen can show
"1–25 of 180" and know whether another page follows. Page size defaults to 25
and is capped at 200 — without a ceiling a caller can ask for every row and
turn a list into a denial of service.

| Endpoint | Filters |
|----------|---------|
| `/admin/members` | `q` (name, email), `team_id`, `role`, `color_status`, `is_active` |
| `/admin/tyfcb` | `q` (either party), `status`, `team_id`, `from`, `to` |
| `/admin/visitors` | `q` (guest, contact, inviter), `status`, `team_id`, `converted` |

Each filter is built by a shared clause builder that binds every value as a
parameter, so a predicate can never be assembled by string concatenation.

## Hardening

- Login is rate limited to ten attempts per minute per IP, and an unknown
  email still pays the cost of a bcrypt comparison so response time cannot be
  used to enumerate accounts.
- Request bodies are capped at 1 MB; every endpoint takes a small JSON object.
- Responses carry `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`,
  `Permissions-Policy` and `Cache-Control: no-store`. HSTS is left to the TLS
  terminator, which knows whether the deployment is actually HTTPS.
- Expired sessions are swept hourly rather than accumulating for the season.
- The dashboard's independent reads run concurrently, and raffle tickets are
  rebuilt set-at-a-time in SQL instead of a query per member.

## Gates

```bash
bash scripts/qa.sh              # gofmt, go vet, tsc, oxlint
bash scripts/test.sh            # go test + both production builds
bash scripts/security-check.sh  # tracked-secret scan + dependency audit
```

CI runs the same checks on every push (`.gitlab-ci.yml`), and deploy only runs
once all three pass.

Tests cover two layers. `internal/domain` pins the rules as pure functions —
bands, pair penalty, event multipliers, badge thresholds. `internal/usecase`
pins the orchestration against in-memory repositories: which ledger rows a
verification writes, that a reversal nets to zero, that a boosted milestone is
given back in full, and that a losing optimistic-lock update awards nothing.

## Documents

- [Product Requirements](./docs/product/PRD.md) — scoring spec and event bank
- [Project Brief](./docs/project-brief.md)
- [Conventions](./docs/CONVENTIONS.md)
