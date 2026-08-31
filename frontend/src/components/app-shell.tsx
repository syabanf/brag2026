import { useEffect, useRef, useState, type ReactNode } from "react";
import { Link, NavLink, useLocation, useNavigate } from "react-router-dom";
import {
  BarChart3,
  ChevronDown,
  Gift,
  History,
  Home,
  KeyRound,
  LogOut,
  Plus,
  Shield,
  Trophy,
  User as UserIcon,
  Zap,
} from "lucide-react";
import { useAuth } from "../lib/use-auth";
import { QuickTour, TourButton } from "./quick-tour";
import { NotificationBell } from "./notification-bell";
import { initials } from "../lib/format";

/**
 * The shell is three layouts, not one that stretches.
 *
 *   phone   a compact top bar and a bottom tab row, thumb-first
 *   tablet  an icon rail down the left, no bottom bar
 *   desktop a full sidebar with labels, and a slim toolbar for page actions
 *
 * The previous version put a floating pill at the bottom of every screen
 * size, which is a phone pattern sitting in the middle of a 1400px window —
 * far from the pointer and covering content it did not need to.
 */

type NavItem = {
  to: string;
  label: string;
  icon: typeof Home;
  end?: boolean;
  /**
   * What to call this where the column is only 5.5rem wide. Without it
   * "Leaderboard" arrives as "Leaderbo…", which names nothing.
   */
  short?: string;
};

/** The five destinations that carry the competition. */
const primaryNav: NavItem[] = [
  { to: "/", label: "Dashboard", icon: Home, end: true },
  { to: "/leaderboard", label: "Leaderboard", icon: BarChart3, short: "Klasemen" },
  { to: "/booster", label: "Booster", icon: Zap },
  { to: "/prizes", label: "Hadiah", icon: Trophy },
];

/** Reachable from the sidebar on a large screen, from the menu on a small one. */
const secondaryNav: NavItem[] = [
  { to: "/awards", label: "Badge", icon: Gift },
  { to: "/activity", label: "Aktivitas", icon: BarChart3 },
  { to: "/history", label: "Riwayat", icon: History },
];

export function AppShell({ children }: { children: ReactNode }) {
  return (
    <div className="min-h-dvh md:pl-[5.5rem] lg:pl-64">
      <SideNav />
      <TopBar />

      {/* The bottom padding only exists where the tab bar does. */}
      <main className="mx-auto w-full max-w-6xl px-4 pb-28 pt-4 sm:px-6 md:pb-10 lg:px-8">
        {children}
      </main>

      <TabBar />
      <QuickTour />
    </div>
  );
}

// ── the rail and the sidebar are one component ────────────────────────────

/**
 * A rail at tablet width and a labelled sidebar on a laptop. One component
 * rather than two, because everything except the label is identical and a
 * second copy would be a second place to forget a route.
 */
function SideNav() {
  const { user } = useAuth();
  const adminPath = user?.role === "admin" ? "/admin" : "/captain";
  const isStaff = user?.role === "admin" || user?.role === "captain";

  return (
    <aside className="fixed inset-y-0 left-0 z-30 hidden w-[5.5rem] flex-col border-r border-black/[0.06] bg-white md:flex lg:w-64">
      <Link
        to="/"
        aria-label="BRAG dashboard"
        className="flex h-16 shrink-0 items-center justify-center border-b border-black/[0.06] lg:justify-start lg:px-5"
      >
        {/* The rail has no room for a wordmark, so it carries a monogram. */}
        <span className="grid h-9 w-9 place-items-center rounded-xl bg-brand-600 text-sm font-black text-white lg:hidden">
          BR
        </span>
        <span className="hidden min-w-0 lg:block">
          <span className="block text-lg font-black leading-none text-brand-600">BRAG 2026</span>
          <span className="mt-1 block text-[0.6rem] font-bold uppercase tracking-[0.14em] text-brand-700">
            BNI Grow Challenge
          </span>
        </span>
      </Link>

      <div className="flex min-h-0 flex-1 flex-col gap-1 overflow-y-auto px-3 py-4 lg:px-3">
        <NavLink
          to="/submit"
          className="mb-2 flex flex-col items-center gap-1 rounded-2xl bg-brand-600 px-1 py-2.5 font-bold text-white transition hover:bg-brand-700 active:scale-[0.98] lg:flex-row lg:gap-2 lg:px-4 lg:py-3"
        >
          <Plus className="h-5 w-5 shrink-0" />
          <span className="text-[0.6rem] leading-tight lg:text-sm">Kontribusi</span>
        </NavLink>

        {primaryNav.map((item) => (
          <RailLink key={item.to} item={item} />
        ))}

        <p className="mt-5 hidden px-3 text-[0.62rem] font-bold uppercase tracking-[0.14em] text-muted lg:block">
          Lainnya
        </p>
        <span className="mt-4 h-px bg-black/[0.06] lg:hidden" />

        {secondaryNav.map((item) => (
          <RailLink key={item.to} item={item} />
        ))}

        {isStaff && (
          <>
            <p className="mt-5 hidden px-3 text-[0.62rem] font-bold uppercase tracking-[0.14em] text-muted lg:block">
              Pengelolaan
            </p>
            <span className="mt-4 h-px bg-black/[0.06] lg:hidden" />
            <RailLink
              item={{
                to: adminPath,
                label: user?.role === "admin" ? "Admin" : "Kapten",
                icon: Shield,
              }}
            />
          </>
        )}
      </div>

      {/* The account sits at the foot of the sidebar the way a desktop app
          puts it — out of the way, always reachable. */}
      <div className="shrink-0 border-t border-black/[0.06] p-3">
        <UserMenu placement="up" />
      </div>
    </aside>
  );
}

function RailLink({ item }: { item: NavItem }) {
  const { to, label, icon: Icon, end, short } = item;

  return (
    <NavLink
      to={to}
      end={end}
      className={({ isActive }) =>
        `flex rounded-2xl transition
         flex-col items-center gap-1 px-1 py-2
         lg:flex-row lg:gap-3 lg:px-4 lg:py-2.5 ${
           isActive
             ? "bg-brand-50 text-brand-700"
             : "text-muted hover:bg-black/[0.03] hover:text-ink"
         }`
      }
    >
      {({ isActive }) => (
        <>
          <Icon className="h-[1.15rem] w-[1.15rem] shrink-0" strokeWidth={isActive ? 2.4 : 2} />
          {/* The rail stacks a small label under the icon — icons alone are
              a puzzle the first time. The sidebar has room to set it beside. */}
          <span className="w-full truncate text-center text-[0.6rem] font-bold leading-tight lg:hidden">
            {short ?? label}
          </span>
          <span className="hidden text-sm font-bold lg:inline">{label}</span>
        </>
      )}
    </NavLink>
  );
}

// ── the bar across the top ────────────────────────────────────────────────

/**
 * On a phone this is the app's identity and its actions. On larger screens
 * the sidebar already carries the identity, so this becomes a slim toolbar:
 * where you are, and what you can do here.
 */
function TopBar() {
  const location = useLocation();
  const title = titleFor(location.pathname);

  return (
    <header className="sticky top-0 z-20 border-b border-black/[0.06] bg-white/75 backdrop-blur-xl">
      <div className="mx-auto flex h-16 w-full max-w-6xl items-center gap-3 px-4 sm:px-6 lg:px-8">
        <Link to="/" className="min-w-0 leading-none md:hidden" aria-label="BRAG dashboard">
          <span className="block text-xl font-black text-brand-600">BRAG 2026</span>
          <span className="mt-0.5 block text-[0.58rem] font-bold uppercase tracking-[0.12em] text-brand-700">
            BNI Grow Challenge
          </span>
        </Link>

        {/* Naming the screen is what makes a sidebar app feel oriented; on a
            phone the heading below already does that, so it stays hidden. */}
        <h1 className="hidden min-w-0 truncate text-base font-black text-ink md:block">{title}</h1>

        <div className="ml-auto flex shrink-0 items-center gap-2">
          <TourButton />
          <NotificationBell />
          {/* Tablet has a rail with no room for an account control, so the
              menu lives up here until the sidebar is wide enough to hold it. */}
          <span className="lg:hidden">
            <UserMenu placement="down" />
          </span>
        </div>
      </div>
    </header>
  );
}

/** The current screen's name, for the toolbar. */
function titleFor(pathname: string) {
  const known: [string, string][] = [
    ["/leaderboard", "Leaderboard"],
    ["/submit", "Kontribusi"],
    ["/booster", "Booster"],
    ["/prizes", "Hadiah"],
    ["/awards", "Badge & Penghargaan"],
    ["/activity", "Aktivitas"],
    ["/history", "Riwayat"],
    ["/profile", "Profil"],
    ["/api-keys", "Kunci API"],
    ["/admin", "Admin"],
    ["/captain", "Tim Saya"],
  ];

  // Longest prefix wins, so /admin/members resolves to Admin rather than to
  // whatever happens to be listed first.
  const match = known
    .filter(([prefix]) => pathname === prefix || pathname.startsWith(prefix + "/"))
    .sort((a, b) => b[0].length - a[0].length)[0];

  return match ? match[1] : "Dashboard";
}

// ── the account menu, shared by both places it appears ────────────────────

function UserMenu({ placement }: { placement: "up" | "down" }) {
  const { user, signOut } = useAuth();
  const navigate = useNavigate();
  const [open, setOpen] = useState(false);
  const box = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;

    function onPointerDown(e: MouseEvent) {
      if (box.current && !box.current.contains(e.target as Node)) setOpen(false);
    }
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") setOpen(false);
    }

    document.addEventListener("mousedown", onPointerDown);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("mousedown", onPointerDown);
      document.removeEventListener("keydown", onKey);
    };
  }, [open]);

  async function handleSignOut() {
    await signOut();
    navigate("/login", { replace: true });
  }

  const isStaff = user?.role === "admin" || user?.role === "captain";

  const items = [
    { to: "/profile", label: "Profil", icon: UserIcon },
    { to: "/api-keys", label: "Kunci API", icon: KeyRound },
  ];

  return (
    <div ref={box} className="relative">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        aria-expanded={open}
        aria-label="Menu akun"
        className={
          placement === "up"
            ? "flex w-full items-center gap-2.5 rounded-2xl p-2 transition hover:bg-black/[0.03] lg:gap-3"
            : "grid h-11 w-11 place-items-center rounded-full border-2 border-brand-600 text-sm font-black text-brand-600 transition hover:bg-brand-50"
        }
      >
        {placement === "up" ? (
          <>
            <span className="grid h-9 w-9 shrink-0 place-items-center rounded-full border-2 border-brand-600 text-xs font-black text-brand-600">
              {initials(user?.full_name)}
            </span>
            <span className="hidden min-w-0 flex-1 text-left lg:block">
              <span className="block truncate text-sm font-bold text-ink">{user?.full_name}</span>
              <span className="block truncate text-xs text-muted">{roleLabel(user?.role)}</span>
            </span>
            <ChevronDown className="hidden h-4 w-4 shrink-0 text-muted lg:block" />
          </>
        ) : (
          initials(user?.full_name)
        )}
      </button>

      {open && (
        <div
          role="menu"
          className={`absolute z-40 w-56 overflow-hidden rounded-2xl border border-black/[0.07] bg-white shadow-xl ${
            placement === "up" ? "bottom-full left-0 mb-2" : "right-0 top-full mt-2"
          }`}
        >
          <div className="border-b border-black/[0.06] px-4 py-3">
            <p className="truncate text-sm font-bold text-ink">{user?.full_name}</p>
            <p className="truncate text-xs text-muted">{user?.email}</p>
          </div>

          {/* A phone has no rail and no sidebar, so without this an admin on
              their phone has no way into the admin area at all. The tablet
              rail already carries it, hence md:hidden. */}
          {isStaff && (
            <div className="md:hidden">
              <MenuLink
                to={user?.role === "admin" ? "/admin" : "/captain"}
                label={user?.role === "admin" ? "Panel Admin" : "Tim Saya"}
                icon={Shield}
                onGo={() => setOpen(false)}
              />
              <span className="block h-px bg-black/[0.06]" />
            </div>
          )}

          {/* On a phone and a tablet these are not in a sidebar, so the menu
              is the only way to reach them. */}
          <div className="lg:hidden">
            {secondaryNav.map(({ to, label, icon: Icon }) => (
              <MenuLink key={to} to={to} label={label} icon={Icon} onGo={() => setOpen(false)} />
            ))}
            <span className="block h-px bg-black/[0.06]" />
          </div>

          {items.map(({ to, label, icon: Icon }) => (
            <MenuLink key={to} to={to} label={label} icon={Icon} onGo={() => setOpen(false)} />
          ))}

          <button
            type="button"
            onClick={handleSignOut}
            className="flex w-full items-center gap-2.5 border-t border-black/[0.06] px-4 py-3 text-sm font-semibold text-brand-600 transition hover:bg-brand-50"
          >
            <LogOut className="h-4 w-4" />
            Keluar
          </button>
        </div>
      )}
    </div>
  );
}

function MenuLink({
  to,
  label,
  icon: Icon,
  onGo,
}: {
  to: string;
  label: string;
  icon: typeof Home;
  onGo: () => void;
}) {
  return (
    <Link
      to={to}
      onClick={onGo}
      role="menuitem"
      className="flex items-center gap-2.5 px-4 py-3 text-sm font-semibold text-ink transition hover:bg-brand-50"
    >
      <Icon className="h-4 w-4 shrink-0 text-muted" />
      {label}
    </Link>
  );
}

function roleLabel(role?: string) {
  if (role === "admin" || role === "superadmin") return "Growth Coordinator";
  if (role === "captain") return "Kapten Tim";
  return "Member";
}

// ── the phone's tab bar ───────────────────────────────────────────────────

/**
 * Five targets across the thumb's reach, with the contribute action raised in
 * the middle. Only on a phone: at tablet width the rail takes over, and a bar
 * at the bottom of a laptop window is a long way from the pointer.
 */
function TabBar() {
  const tabs: (NavItem & { primary?: boolean })[] = [
    primaryNav[0],
    primaryNav[1],
    { to: "/submit", label: "Kontribusi", icon: Plus, primary: true },
    primaryNav[2],
    primaryNav[3],
  ];

  return (
    <nav className="safe-bottom fixed inset-x-0 bottom-0 z-30 border-t border-black/[0.06] bg-white/90 px-2 pt-1.5 backdrop-blur-xl md:hidden">
      <div className="mx-auto grid max-w-md grid-cols-5 items-end gap-1">
        {tabs.map(({ to, label, icon: Icon, primary, end, short }) =>
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
                  <span className="w-full truncate">{short ?? label}</span>
                </>
              )}
            </NavLink>
          ),
        )}
      </div>
    </nav>
  );
}
