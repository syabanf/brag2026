import Link from "next/link";
import { getCurrentUser } from "@/lib/local-auth";
import { getDemoRole, isDemoActive } from "@/lib/demo-session";
import { NotificationBell } from "./notification-bell";
import { ProfileMenu } from "./profile-menu";
import { DesktopNav } from "./desktop-nav";
import { MobileNav } from "./mobile-nav";
import { QuickTour } from "./quick-tour";
import { TourButton } from "./tour-button";
import { DemoBadge } from "./demo-badge";

export async function AppShell({ children }: { children: React.ReactNode }) {
  const user = await getCurrentUser();
  const isAdmin = user?.role === "admin";
  const isCaptain = user?.role === "captain";

  const demo = await isDemoActive();
  const demoRole = await getDemoRole();

  return (
    <div className="mx-auto flex min-h-screen w-full max-w-6xl flex-col px-4 pb-28 pt-5 sm:px-6 lg:px-8 lg:pb-28">
      <header className="mb-5 flex items-center justify-between gap-3">
        <Link href="/" className="min-w-0 leading-none" aria-label="BRAG dashboard">
          <span className="block text-2xl font-black tracking-normal text-brand-600 sm:text-4xl">
            BRAG 2026
          </span>
          <span className="mt-1 block text-[0.6rem] font-bold uppercase tracking-[0.12em] text-brand-700 sm:text-[0.68rem] sm:tracking-[0.18em]">
            <span className="sm:hidden">BNI Grow Challenge</span>
            <span className="hidden sm:inline">BNI Grow Annual Challenge</span>
          </span>
        </Link>

        <div className="flex shrink-0 items-center gap-1.5 sm:gap-3">
          <div className="hidden text-right lg:block">
            <p className="text-sm font-semibold text-ink">Season 2026</p>
            <p className="text-xs text-muted">Week 2 of 12</p>
          </div>
          <TourButton />
          <NotificationBell />
          <ProfileMenu
            initials={
              user?.full_name
                ? user.full_name.split(" ").map((n) => n[0]).join("").slice(0, 2).toUpperCase()
                : "?"
            }
            isAdmin={isAdmin}
            isCaptain={isCaptain}
          />
        </div>
      </header>

      {demo && <DemoBadge role={demoRole} />}

      <main className="flex-1">{children}</main>

      <DesktopNav />

      <MobileNav />

      <QuickTour />
    </div>
  );
}
