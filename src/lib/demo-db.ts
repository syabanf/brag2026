import { readFile } from "node:fs/promises";
import path from "node:path";
import type { QueryResultRow } from "pg";
import { PGlite } from "@electric-sql/pglite";
import { pgcrypto } from "@electric-sql/pglite/contrib/pgcrypto";

// Applied in order. Demo seed runs last so it can reference seeded members/teams.
const MIGRATIONS = [
  "db/local/001_initial.sql",
  "db/local/002_seed_members.sql",
  "db/local/003_booster_events.sql",
  "db/local/005_captain_role.sql",
  "db/demo/001_demo_seed.sql"
];

// Next.js dev reloads modules on every edit; the instance is parked on
// globalThis so the in-memory database survives hot reloads.
const globalForDemo = globalThis as unknown as { __bragDemoDb?: Promise<PGlite> };

async function bootstrap(): Promise<PGlite> {
  const db = await PGlite.create({ extensions: { pgcrypto } });

  for (const file of MIGRATIONS) {
    const sql = await readFile(path.join(process.cwd(), file), "utf8");
    await db.exec(sql);
  }

  return db;
}

export function getDemoDb(): Promise<PGlite> {
  if (!globalForDemo.__bragDemoDb) {
    globalForDemo.__bragDemoDb = bootstrap();
  }

  return globalForDemo.__bragDemoDb;
}

export async function demoQuery<T extends QueryResultRow>(
  text: string,
  values: unknown[] = []
) {
  const db = await getDemoDb();
  const result = await db.query<T>(text, values);

  return {
    rows: result.rows,
    rowCount: result.rows.length > 0 ? result.rows.length : result.affectedRows ?? 0
  };
}
