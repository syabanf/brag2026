import { cookies } from "next/headers";
import { DEMO_COOKIE, DEMO_ROLE_COOKIE, isDemoAvailable, isDemoRole } from "@/lib/demo";

const ONE_DAY = 60 * 60 * 24;

export async function POST(request: Request) {
  if (!isDemoAvailable()) {
    return Response.json({ error: "Demo mode is disabled" }, { status: 403 });
  }

  const { role } = (await request.json().catch(() => ({}))) as { role?: string };
  const store = await cookies();

  store.set(DEMO_COOKIE, "1", { httpOnly: true, path: "/", sameSite: "lax", maxAge: ONE_DAY });
  store.set(DEMO_ROLE_COOKIE, isDemoRole(role) ? role : "admin", {
    httpOnly: true,
    path: "/",
    sameSite: "lax",
    maxAge: ONE_DAY
  });

  return Response.json({ ok: true });
}

export async function DELETE() {
  const store = await cookies();
  store.delete(DEMO_COOKIE);
  store.delete(DEMO_ROLE_COOKIE);
  return Response.json({ ok: true });
}
