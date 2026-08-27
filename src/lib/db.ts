import { Pool, type QueryResultRow } from "pg";
import { isDemoActive } from "@/lib/demo-session";

let pool: Pool | null = null;

export function getDb() {
  if (!pool) {
    pool = new Pool({
      connectionString: process.env.DATABASE_URL ?? "postgresql:///brag_dev"
    });
  }

  return pool;
}

export async function query<T extends QueryResultRow>(
  text: string,
  values: unknown[] = []
) {
  // Demo mode swaps PostgreSQL for an in-memory PGlite instance running the
  // same schema, so every caller keeps working untouched. Imported lazily so
  // the WASM bundle never loads in a normal run.
  if (await isDemoActive()) {
    const { demoQuery } = await import("@/lib/demo-db");
    return demoQuery<T>(text, values);
  }

  return getDb().query<T>(text, values);
}
