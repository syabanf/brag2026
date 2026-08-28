import { createContext } from "react";
import type { Member, User } from "./types";

export type AuthState = {
  user: User | null;
  member: Member | null;
  loading: boolean;
  signIn: (email: string, password: string) => Promise<void>;
  signOut: () => Promise<void>;
  refresh: () => Promise<void>;
};

/**
 * Kept out of the provider module so that file exports a component and
 * nothing else, which is what React Fast Refresh needs to swap it cleanly.
 */
export const AuthContext = createContext<AuthState | null>(null);
