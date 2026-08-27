import crypto from "node:crypto";
import { cookies } from "next/headers";
import { query } from "@/lib/db";
import { DEMO_ACCOUNTS } from "@/lib/demo";
import { getDemoRole, isDemoActive } from "@/lib/demo-session";

export const SESSION_COOKIE = "brag_session";

export type LocalUser = {
  id: string;
  email: string;
  full_name: string;
  role: "member" | "captain" | "admin" | "super_admin";
};

function hashToken(token: string) {
  return crypto.createHash("sha256").update(token).digest("hex");
}

export async function signInWithPassword(email: string, password: string) {
  const result = await query<LocalUser>(
    `select id, email, full_name, role
     from app_users
     where lower(email) = lower($1)
       and password_hash = crypt($2, password_hash)
     limit 1`,
    [email, password]
  );

  const user = result.rows[0];
  if (!user) {
    return { error: "Invalid email or password" };
  }

  const token = crypto.randomBytes(32).toString("base64url");
  await query(
    `insert into user_sessions (user_id, token_hash, expires_at)
     values ($1, $2, now() + interval '30 days')`,
    [user.id, hashToken(token)]
  );

  const cookieStore = await cookies();
  cookieStore.set(SESSION_COOKIE, token, {
    httpOnly: true,
    maxAge: 60 * 60 * 24 * 30,
    path: "/",
    sameSite: "lax",
    secure: process.env.NODE_ENV === "production"
  });

  return { user };
}

export async function getCurrentUser() {
  const cookieStore = await cookies();

  // Demo mode has no real sessions: the persona cookie resolves straight to a
  // seeded account so the app is browsable without signing in.
  if (await isDemoActive()) {
    const role = await getDemoRole();

    const demo = await query<LocalUser>(
      `select id, email, full_name, role from app_users where lower(email) = lower($1) limit 1`,
      [DEMO_ACCOUNTS[role].email]
    );

    return demo.rows[0] ?? null;
  }

  const token = cookieStore.get(SESSION_COOKIE)?.value;

  if (!token) {
    return null;
  }

  const result = await query<LocalUser>(
    `select u.id, u.email, u.full_name, u.role
     from user_sessions s
     join app_users u on u.id = s.user_id
     where s.token_hash = $1
       and s.expires_at > now()
     limit 1`,
    [hashToken(token)]
  );

  return result.rows[0] ?? null;
}

export async function signOut() {
  const cookieStore = await cookies();
  const token = cookieStore.get(SESSION_COOKIE)?.value;

  if (token) {
    await query("delete from user_sessions where token_hash = $1", [
      hashToken(token)
    ]);
  }

  cookieStore.delete(SESSION_COOKIE);
}
