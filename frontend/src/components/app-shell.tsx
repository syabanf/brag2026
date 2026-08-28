import { useState, type ReactNode } from "react";
import { Link, NavLink, useNavigate } from "react-router-dom";
import {
  BarChart3,
  Home,
  LogOut,
  Plus,
  Shield,
  Trophy,
  User as UserIcon,
  Users,
  Zap,
} from "lucide-react";
import { useAuth } from "../lib/use-auth";
import { QuickTour, TourButton } from "./quick-tour";
import { initials } from "../lib/format";

const navItems = [
  { to: "/", label: "Dashboard", icon: Home, end: true },
  { to: "/leaderboard", label: "Leaderboard", icon: BarChart3 },
  { to: "/submit", label: "Contribute", icon: Plus, primary: true },
  { to: "/booster", label: "Booster", icon: Zap },
  { to: "/prizes", label: "Hadiah", icon: Trophy },
];

export function AppShell({ children }: { children: ReactNode }) {
  const { user, signOut } = useAuth();
  const navigate = useNavigate();
  const [menuOpen, setMenuOpen] = useState(false);

  async function handleSignOut() {
    await signOut();
    navigate("/login", { replace: true });
  }

  return (
    <div className="mx-auto flex min-h-screen w-full max-w-6xl flex-col px-4 pb-28 pt-5 sm:px-6 lg:px-8">
      <header className="mb-5 flex items-center justify-between gap-3">
        <Link to="/" className="min-w-0 leading-none" aria-label="BRAG dashboard">
          <span className="block text-2xl font-black text-brand-600 sm:text-4xl">BRAG 2026</span>
          <span className="mt-1 block text-[0.6rem] font-bold uppercase tracking-[0.12em] text-brand-700 sm:text-[0.68rem] sm:tracking-[0.18em]">
            <span className="sm:hidden">BNI Grow Challenge</span>
            <span className="hidden sm:inline">BNI Grow Annual Challenge</span>
          </span>
        </Link>

        <div className="relative flex shrink-0 items-center gap-2">
          <TourButton />

          {(user?.role === "admin" || user?.role === "captain") && (
            <Link
              to={user.role === "admin" ? "/admin/members" : "/captain"}
              className="flex min-h-11 items-center gap-1.5 rounded-full border border-brand-100 bg-brand-50 px-3 text-sm font-bold text-brand-700 transition hover:bg-brand-100"
            >
              <Shield className="h-[1.05rem] w-[1.05rem]" />
              <span className="hidden sm:inline">
                {user.role === "admin" ? "Admin" : "Kapten"}
              </span>
            </Link>
          )}

          <button
            type="button"
            onClick={() => setMenuOpen((v) => !v)}
            aria-expanded={menuOpen}
            aria-label="Menu profil"
            className="grid h-11 w-11 place-items-center rounded-full border-2 border-brand-600 text-sm font-black text-brand-600 transition hover:bg-brand-50"
          >
            {initials(user?.full_name)}
          </button>

          {menuOpen && (
            <>
              <div
                className="fixed inset-0 z-30"
                aria-hidden
                onClick={() => setMenuOpen(false)}
              />
              <div className="absolute right-0 top-full z-40 mt-2 w-52 overflow-hidden rounded-2xl border border-brand-100 bg-white shadow-lift">
                <div className="border-b border-brand-50 px-4 py-3">
                  <p className="truncate text-sm font-bold text-ink">{user?.full_name}</p>
                  <p className="truncate text-xs text-muted">{user?.email}</p>
                </div>
                <Link
                  to="/profile"
                  onClick={() => setMenuOpen(false)}
                  className="flex items-center gap-2 px-4 py-3 text-sm font-semibold text-ink transition hover:bg-brand-50"
                >
                  <UserIcon className="h-4 w-4" />
                  Profil
                </Link>
                <Link
                  to="/history"
                  onClick={() => setMenuOpen(false)}
                  className="flex items-center gap-2 px-4 py-3 text-sm font-semibold text-ink transition hover:bg-brand-50"
                >
                  <BarChart3 className="h-4 w-4" />
                  Riwayat
                </Link>
                <button
                  type="button"
                  onClick={handleSignOut}
                  className="flex w-full items-center gap-2 border-t border-brand-50 px-4 py-3 text-sm font-semibold text-brand-600 transition hover:bg-brand-50"
                >
                  <LogOut className="h-4 w-4" />
                  Keluar
                </button>
              </div>
            </>
          )}
        </div>
      </header>

      <main className="flex-1">{children}</main>

      <DesktopNav />
      <MobileNav />
      <QuickTour />
    </div>
  );
}

function DesktopNav() {
  return (
    <nav className="fixed bottom-6 left-1/2 z-20 hidden -translate-x-1/2 lg:block">
      <div className="flex items-center gap-1 rounded-full border border-brand-100 bg-white px-3 py-2 shadow-[0_8px_48px_rgba(0,0,0,0.18)]">
        {navItems.map(({ to, label, icon: Icon, primary, end }) =>
          primary ? (
            <NavLink
              key={to}
              to={to}
              aria-label={label}
              className="mx-1 flex h-12 w-12 items-center justify-center rounded-full bg-brand-600 text-white shadow-md transition hover:bg-brand-700 active:scale-95"
            >
              <Icon className="h-5 w-5" />
            </NavLink>
          ) : (
            <NavLink
              key={to}
              to={to}
              end={end}
              className={({ isActive }) =>
                `flex items-center gap-2 rounded-full px-4 py-2.5 text-sm font-semibold transition ${
                  isActive ? "bg-brand-600 text-white" : "text-muted hover:bg-brand-50 hover:text-brand-600"
                }`
              }
            >
              <Icon className="h-[1.1rem] w-[1.1rem]" />
              <span>{label}</span>
            </NavLink>
          ),
        )}
      </div>
    </nav>
  );
}

function MobileNav() {
  return (
    <nav className="safe-bottom fixed inset-x-0 bottom-0 z-20 border-t border-black/[0.06] bg-white/90 px-2 pt-1.5 backdrop-blur-xl lg:hidden">
      <div className="mx-auto grid max-w-md grid-cols-5 items-end gap-1">
        {navItems.map(({ to, label, icon: Icon, primary, end }) =>
          primary ? (
            <NavLink
              key={to}
              to={to}
              aria-label={label}
              className="-mt-6 flex h-14 w-14 items-center justify-center justify-self-center rounded-2xl bg-brand-600 text-white shadow-[0_8px_20px_-6px_rgba(200,16,46,0.6)] transition active:scale-95"
            >
              <Icon className="h-6 w-6" />
            </NavLink>
          ) : (
            <NavLink
              key={to}
              to={to}
              end={end}
              className={({ isActive }) =>
                `flex min-h-14 flex-col items-center justify-center gap-1 rounded-xl px-0.5 text-center text-[0.6rem] font-bold leading-tight transition ${
                  isActive ? "text-brand-600" : "text-muted"
                }`
              }
            >
              {({ isActive }) => (
                <>
                  <Icon className="h-5 w-5 shrink-0" strokeWidth={isActive ? 2.5 : 2} />
                  <span className="w-full truncate">{label}</span>
                </>
              )}
            </NavLink>
          ),
        )}
      </div>
    </nav>
  );
}

export { Users };
