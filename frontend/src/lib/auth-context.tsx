import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { api, ApiError } from "./api";
import type { Member, User } from "./types";

type AuthState = {
  user: User | null;
  member: Member | null;
  loading: boolean;
  signIn: (email: string, password: string) => Promise<void>;
  signOut: () => Promise<void>;
  refresh: () => Promise<void>;
};

const AuthContext = createContext<AuthState | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [member, setMember] = useState<Member | null>(null);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    try {
      const { user, member } = await api.auth.me();
      setUser(user);
      setMember(member);
    } catch (err) {
      // A 401 simply means nobody is signed in; anything else is worth surfacing.
      if (!(err instanceof ApiError) || err.status !== 401) {
        console.error("session check failed", err);
      }
      setUser(null);
      setMember(null);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const signIn = useCallback(
    async (email: string, password: string) => {
      await api.auth.login(email, password);
      await refresh();
    },
    [refresh],
  );

  const signOut = useCallback(async () => {
    await api.auth.logout();
    setUser(null);
    setMember(null);
  }, []);

  const value = useMemo(
    () => ({ user, member, loading, signIn, signOut, refresh }),
    [user, member, loading, signIn, signOut, refresh],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used inside <AuthProvider>");
  }
  return ctx;
}
