import { cookies } from "next/headers";
import {
  DEMO_COOKIE,
  DEMO_ROLE_COOKIE,
  isDemoAvailable,
  isDemoRole,
  type DemoRole
} from "@/lib/demo";

/** Whether THIS visitor chose demo mode on the landing page. */
export async function isDemoActive() {
  if (!isDemoAvailable()) return false;

  try {
    const store = await cookies();
    return store.get(DEMO_COOKIE)?.value === "1";
  } catch {
    // Outside a request scope (build-time evaluation) there is no visitor.
    return false;
  }
}

export async function getDemoRole(): Promise<DemoRole> {
  try {
    const store = await cookies();
    const value = store.get(DEMO_ROLE_COOKIE)?.value;
    return isDemoRole(value) ? value : "admin";
  } catch {
    return "admin";
  }
}
