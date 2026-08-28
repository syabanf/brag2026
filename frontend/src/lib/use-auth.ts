import { useContext } from "react";
import { AuthContext } from "./auth-store";

/**
 * Lives apart from the provider so the context module exports only a
 * component, which is what React Fast Refresh needs to swap it cleanly.
 */
export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used inside <AuthProvider>");
  }
  return ctx;
}
