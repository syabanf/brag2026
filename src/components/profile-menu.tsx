"use client";

import { Banknote, Briefcase, KeyRound, LogOut, Shield, SlidersHorizontal, UserCheck, UserCircle, Users, Zap, UsersRound } from "lucide-react";
import Link from "next/link";
import { useEffect, useRef, useState } from "react";
import { logout } from "@/app/login/actions";

const adminLinks = [
  { href: "/admin/tyfcb",   label: "Verifikasi TYFCB",   icon: Banknote },
  { href: "/admin/visitors", label: "Kelola Visitor",     icon: UserCheck },
  { href: "/admin/members", label: "Kelola Member",       icon: Users },
  { href: "/admin/teams",   label: "Kelola Team",         icon: UsersRound },
  { href: "/admin/classifications", label: "Klasifikasi Bisnis", icon: Briefcase },
  { href: "/admin/booster", label: "Kelola Booster",      icon: Zap },
  { href: "/admin/scoring", label: "Aturan Skor",         icon: SlidersHorizontal },
];

export function ProfileMenu({
  initials,
  isAdmin,
  isCaptain,
}: {
  initials: string;
  isAdmin: boolean;
  isCaptain?: boolean;
}) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, []);

  return (
    <div className="relative" ref={ref}>
      <button
        onClick={() => setOpen((v) => !v)}
        className="grid h-11 w-11 place-items-center rounded-full border-2 border-brand-600 bg-white text-sm font-bold text-brand-700 transition hover:bg-brand-50"
        aria-label="Profile menu"
      >
        {initials}
      </button>

      {open && (
        <div className="absolute right-0 top-13 z-50 mt-2 w-56 overflow-hidden rounded-2xl border border-brand-100 bg-white shadow-lift">
          {/* Admin section */}
          {isAdmin && (
            <>
              <div className="border-b border-brand-50 px-4 py-2">
                <p className="text-[0.68rem] font-bold uppercase tracking-[0.14em] text-brand-700">
                  Admin Area
                </p>
              </div>
              {adminLinks.map((link) => {
                const Icon = link.icon;
                return (
                  <Link
                    key={link.href}
                    href={link.href}
                    onClick={() => setOpen(false)}
                    className="flex items-center gap-3 px-4 py-3 text-sm font-semibold text-ink hover:bg-brand-50"
                  >
                    <Icon className="h-4 w-4 text-brand-600" />
                    {link.label}
                  </Link>
                );
              })}
              <div className="border-t border-brand-50" />
            </>
          )}

          {/* Captain section */}
          {isCaptain && (
            <>
              <div className="border-b border-brand-50 px-4 py-2">
                <p className="text-[0.68rem] font-bold uppercase tracking-[0.14em] text-amber-700">
                  Kapten Tim
                </p>
              </div>
              <Link
                href="/captain"
                onClick={() => setOpen(false)}
                className="flex items-center gap-3 px-4 py-3 text-sm font-semibold text-ink hover:bg-amber-50"
              >
                <Shield className="h-4 w-4 text-amber-600" />
                Panel Kapten
              </Link>
              <div className="border-t border-brand-50" />
            </>
          )}

          {/* Profile & settings */}
          <Link
            href="/profile"
            onClick={() => setOpen(false)}
            className="flex items-center gap-3 px-4 py-3 text-sm font-semibold text-ink hover:bg-brand-50"
          >
            <UserCircle className="h-4 w-4 text-brand-600" />
            Profil & Ganti Password
          </Link>
          <div className="border-t border-brand-50" />

          {/* Logout */}
          <form action={logout}>
            <button
              type="submit"
              className="flex w-full items-center gap-3 px-4 py-3 text-sm font-semibold text-muted hover:bg-brand-50 hover:text-ink"
            >
              <LogOut className="h-4 w-4" />
              Keluar
            </button>
          </form>
        </div>
      )}
    </div>
  );
}
