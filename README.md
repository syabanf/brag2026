# BRAG

BRAG is a member gamification platform for the BNI Grow annual challenge.

The program runs for around three months. Members are grouped by the committee, then each member contribution adds points to their group leaderboard. The main contribution categories are TYFCB, visitor brought, and referral activity. Bonus mechanics such as booster points, special business classifications, and campaign days will be defined later.

## Project Status

Initial Next.js MVP scaffold in progress. The app currently includes the first
member dashboard, login screen, submission form, leaderboard, awards view, admin
verification surface, local PostgreSQL auth, and an initial local database
migration.

## Product Direction

- Mobile-first web app for members.
- Desktop-friendly admin and committee views.
- Visual identity follows the existing BNI Grow Visitor Manager theme.
- Main experience should feel simple, energetic, and competitive.

## Core Users

- Member: logs in, submits contribution activity, tracks personal and group standing.
- Admin or committee: manages members, groups, scoring rules, boosts, verification, and leaderboard.
- Viewer: can see public leaderboard or award highlights if enabled.

## First Documents

- [Project Brief](./docs/project-brief.md)
- [Product Requirements](./docs/product-requirements.md)
- [Gamification Draft](./docs/gamification-draft.md)
- [Architecture Notes](./docs/architecture.md)

## Local Development

`.env` is committed, so a fresh clone runs the demo with no setup:

```bash
npm install
npm run dev
```

Personal overrides go in `.env.local`, which stays gitignored and takes
precedence over `.env`.

### Demo mode (no database required)

Opening the app lands on `/welcome`, where you choose how to enter:

- **Mode Demo** — runs on an in-memory [PGlite](https://pglite.dev) database
  (PostgreSQL compiled to WASM) that applies the real `db/local/` migrations
  plus `db/demo/001_demo_seed.sql`. No PostgreSQL install, no login. The demo
  seeds a season already in flight: 10 teams, 100 members, 180 TYFCB entries,
  60 visitors and a populated score ledger. Switch between the Admin, Captain
  and Member personas from the bar at the top of any page.
- **Masuk dengan akun** — the normal login against your real PostgreSQL.

Demo state lives in a cookie, so the two modes coexist. Set
`NEXT_PUBLIC_DEMO_MODE=false` to hide the demo option entirely (do this in
production). Anything written in demo mode is discarded when the server stops.

```bash
npm install
npm run dev
```

### Quick tour

The compass button in the header starts a 10-step guided tour that walks
across the app and narrates each step. Credentials already live in `.env`:

```bash
ELEVENLABS_API_KEY=...
ELEVENLABS_VOICE_ID=1k39YpzqXZn52BgyLyGO
```

The key is read server-side only and never reaches the browser.

**Heads up:** that voice is an ElevenLabs *professional* library voice, which
free plans cannot use through the API — the call returns HTTP 402 and the tour
narrates with the browser's built-in speech synthesis instead. For real
ElevenLabs audio on a free plan, switch to a premade voice:

```bash
ELEVENLABS_VOICE_ID=EXAVITQu4vr4xnSDxMaL   # Sarah — neutral, clear
```

## Docker

The image is self-contained: demo mode needs no database at all, and the
compose stack adds PostgreSQL with the migrations applied automatically.

```bash
docker compose up --build
```

Then open http://localhost:3000. Both entry paths work:

- **Mode Demo** runs on in-memory PGlite inside the app container — the `db`
  service is not touched.
- **Masuk dengan akun** queries the `db` service, whose volume is initialised
  once from `db/local/*.sql` in filename order.

If port 3000 or 5433 is taken:

```bash
APP_PORT=3200 DB_PORT=5434 docker compose up --build
```

To hide the demo option in a production image (`NEXT_PUBLIC_*` is inlined at
build time, so this is a build argument, not a runtime variable):

```bash
docker build --build-arg NEXT_PUBLIC_DEMO_MODE=false -t brag2026 .
```

The app runs as a non-root user, exposes `/api/health` for the healthcheck,
and ships `db/` so PGlite can apply the migrations at runtime.

### Full local database

For the real thing, create the database and apply the migrations manually
(these are never applied automatically):

```bash
psql -d postgres -c "create database brag_dev;"
for f in db/local/*.sql; do psql -d brag_dev -f "$f"; done
npm run dev
```

Default local superadmin:

- Email: `ilham@wit.id`
- Password: `admin123`

Seeded members use `m<team><member>@brag2026.id` (e.g. `m11@brag2026.id`) with
password `member123`.
