import { Pool, type QueryResultRow } from "pg";

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
  return getDb().query<T>(text, values);
}

/** Runs fn inside a transaction, rolling back if it throws. */
export async function withTransaction<T>(
  fn: (q: <R extends QueryResultRow>(text: string, values?: unknown[]) => Promise<{ rows: R[] }>) => Promise<T>
): Promise<T> {
  const client = await getDb().connect();
  try {
    await client.query("begin");
    const result = await fn((text, values = []) => client.query(text, values));
    await client.query("commit");
    return result;
  } catch (error) {
    await client.query("rollback");
    throw error;
  } finally {
    client.release();
  }
}
