"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { BarChart3, Home, Plus, Users, Zap } from "lucide-react";

const navItems = [
  { href: "/",              label: "Dashboard",   icon: Home },
  { href: "/leaderboard",   label: "Leaderboard", icon: BarChart3 },
  { href: "/submit",        label: "Contribute",  icon: Plus, primary: true },
  { href: "/booster",       label: "Booster",     icon: Zap },
  { href: "/admin/members", label: "Member",      icon: Users }
];

export function MobileNav() {
  const pathname = usePathname();

  return (
    <nav className="safe-bottom fixed inset-x-0 bottom-0 z-20 border-t border-black/[0.06] bg-white/90 px-2 pt-1.5 backdrop-blur-xl lg:hidden">
      <div className="mx-auto grid max-w-md grid-cols-5 items-end gap-1">
        {navItems.map((item) => {
          const Icon = item.icon;
          const active = item.href === "/"
            ? pathname === "/"
            : pathname.startsWith(item.href);

          if (item.primary) {
            return (
              <Link
                key={item.href}
                href={item.href}
                aria-label={item.label}
                className="-mt-6 flex h-14 w-14 items-center justify-center justify-self-center rounded-2xl bg-brand-600 text-white shadow-[0_8px_20px_-6px_rgba(200,16,46,0.6)] transition active:scale-95"
              >
                <Icon className="h-6 w-6" />
              </Link>
            );
          }

          return (
            <Link
              key={item.href}
              href={item.href}
              aria-current={active ? "page" : undefined}
              className={`flex min-h-14 flex-col items-center justify-center gap-1 rounded-xl px-0.5 text-center text-[0.6rem] font-bold leading-tight transition ${
                active ? "text-brand-600" : "text-muted"
              }`}
            >
              <Icon className="h-5 w-5 shrink-0" strokeWidth={active ? 2.5 : 2} />
              <span className="w-full truncate">{item.label}</span>
            </Link>
          );
        })}
      </div>
    </nav>
  );
}
