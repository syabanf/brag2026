# BRAG — Project OS

> Gamification platform for the BNI Grow annual member challenge.

@docs/CONVENTIONS.md @docs/AGENTS.md

---

## Stack

| Key | Value |
|-----|-------|
| Type | Monorepo: Go API + Vite SPA |
| Backend | Go 1.26, clean architecture, chi + pgx |
| Frontend | Vite + React + TypeScript + Tailwind |
| Database | PostgreSQL 17 |
| Auth | bcrypt + HttpOnly session cookie |
| Styling | Tailwind, DM Sans, BNI red `#C8102E` |
| Integration branch | `main` |

**Layering (backend):** dependencies point inwards. `domain` imports nothing
from the other layers; `usecase` depends only on `domain`; `repository` and
`delivery` are interchangeable outer adapters. `cmd/api` is the composition
root and the only place that knows every concrete type.

**Protected paths** (require explicit approval before editing):
- `backend/migrations/` — schema; applied by compose on first boot
- `.env` files — never commit; each service ships a `.env.example`

## Domain Model (Spec v1.0)

**Roles:** `member` | `captain` (input atas nama tim) | `admin` (Growth Coordinator — satu-satunya approver)

**Core tables:**
- `members` — competition profile (user_id, team_id, klasifikasi_id, color_status: merah|kuning|hijau)
- `tyfcb_entries` — TYFCB submissions (giver_id ≠ receiver_id, nilai, status: pending|verified|rejected)
- `visitors` — visitor registrations (inviter_id, status_hadir: terdaftar|hadir|hadir_penuh, is_converted)
- `weekly_events` — 12 event codes, one active per week
- **`score_ledger`** — single source of truth for ALL aggregation (append-only)
- `badges` + `member_badges` — 12 badges, auto-awarded
- `prize_pool` + `raffle_tickets` — two-layer prize system

**Scoring:**
- TYFCB: `computed_score = round(B × P × M)` — Band × Pair Penalty × Event Multiplier
- Visitor: milestone increments (20 → +30 → +100) via admin update
- Team bonus: Full Roster +100/week, Level Up +75/+150
- Flat bonus (NOT multiplied): HIGH_ROLLER +50, ONE_TO_ONE +30, STREAK +40

**NO "referral" category** — spec only has TYFCB + Visitor.

---

## The Front Door — Commands

| Command | Purpose |
|---------|---------|
| `/epic-loop [EPIC-FILE]` | Autonomous epic execution — plan, code, review, test, deploy DEV |
| `/task-work EPIC-XXX [n]` | Implement one task from an epic |
| `/epic-new` | Plan a new epic and write docs/epics/EPIC-XXX.md |
| `/nerve "<task>"` | Three-layer intelligent task execution |
| `/engage <persona> "<task>"` | Activate expert persona then execute |
| `/code-review` | Review current diff for bugs and cleanups |
| `/verify` | Run app and verify a change works end-to-end |

---

## The Nervous System

Three-layer routing for all tasks:

```
L1 SENSE     → Classify task; handle trivial cases directly
L2 COORDINATE → Route to best agent; run gates (QA → Test → Security)
L3 ESCALATE  → Human judgment for ambiguity, conflicts, or retry exhaustion (3+)
```

**Gate scripts:**
- `scripts/qa.sh` — gofmt + go vet + tsc + oxlint
- `scripts/test.sh` — go test + both production builds
- `scripts/security-check.sh` — tracked-secret scan + dependency audit

**Guardrails (always enforced):**
- No `git push --force` to any branch
- No production deployments from autonomous runs
- No SQL `DROP`, `TRUNCATE`, or unguarded `DELETE FROM` (without WHERE)
- No hardcoded secrets — Go reads `os.Getenv`, Vite reads `import.meta.env`
- `score_ledger` is append-only; corrections are opposite-sign rows

---

## Agents

| Agent | Role | Reuse Policy |
|-------|------|--------------|
| `code-agent` | Implements one task; no commits, no scope invention | Project conventions-aware coder |
| `epic-orchestrator` | Drives epic loop across all agents | Swarm coordinator |
| `review-qa-agent` | Checks change against acceptance criteria | QA gate |
| `security-agent` | Runs security-check.sh + checks tenant guards | Security gate |
| `test-agent` | Runs lint, typecheck, build | Test gate |

Prefer ECC/Ruflo pre-built agents for generic tasks. Custom agents above for BRAG-specific conventions only.

---

## Expert Personas

| Activation | Best For |
|------------|---------|
| `/engage startup-mvp "..."` | Build a new feature from scratch end-to-end |
| `/engage frontend-engineer "..."` | Mobile-first UI, Tailwind components |
| `/engage backend-systems "..."` | API design, DB schema, scoring engine |
| `/engage security-audit "..."` | Auth flows, SQL injection checks |
| `/engage codebase-audit "..."` | Understand an unfamiliar part of BRAG |
| `/engage debug-production "..."` | Investigate a live bug |
| `/engage clean-architecture "..."` | Refactor without behavior change |

---

## Epic Lifecycle

```
backlog → on-progress → coding → review → testing → deploying-dev → ready-for-qa → done
                                                                             ↓
                                                                          blocked
```

Epics live in `docs/epics/`. The registry is `docs/epics/README.md`.  
Every code decision is logged in the epic's **Automation Log**.

---

## Key Paths

```
backend/cmd/api/              composition root
backend/internal/domain/      entities, scoring rules, repo interfaces, errors
backend/internal/usecase/     application rules
backend/internal/repository/  PostgreSQL adapters
backend/internal/delivery/    HTTP handlers, routing, middleware
backend/migrations/           schema
frontend/src/lib/api.ts       the single typed contract to the API
frontend/src/pages/           one file per screen
frontend/src/components/      shared UI
docs/product/PRD.md           scoring spec and the 12-event bank
scripts/                      qa, test, security gates
```

---

## Deployment & Security Rules

**Allowed in autonomous runs:**
- All local DB operations
- `go test`, `go build`, `npm run dev`, `npm run build`
- Creating branches, commits, and PRs

**Manual-only (never autonomous):**
- Production deployments
- Database schema migrations (`backend/migrations/*.sql`)
- Changing any `.env` file or secret value
- Force-push to `main`
