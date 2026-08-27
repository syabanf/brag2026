import { cookies } from "next/headers";
import { DEMO_ROLE_COOKIE, isDemoRole } from "@/lib/demo";
import { isDemoActive } from "@/lib/demo-session";

export async function POST(request: Request) {
  if (!(await isDemoActive())) {
    return Response.json({ error: "Demo mode is disabled" }, { status: 403 });
  }

  const { role } = (await request.json()) as { role?: string };

  if (!isDemoRole(role)) {
    return Response.json({ error: "Unknown demo role" }, { status: 400 });
  }

  const cookieStore = await cookies();
  cookieStore.set(DEMO_ROLE_COOKIE, role, {
    httpOnly: true,
    path: "/",
    sameSite: "lax"
  });

  return Response.json({ role });
}
