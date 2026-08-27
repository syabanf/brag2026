"use client";

import { Compass } from "lucide-react";
import { TOUR_EVENT } from "./quick-tour";

export function TourButton({ className = "" }: { className?: string }) {
  return (
    <button
      type="button"
      onClick={() => window.dispatchEvent(new Event(TOUR_EVENT))}
      aria-label="Mulai quick tour"
      className={`flex min-h-11 items-center gap-1.5 rounded-full border border-brand-100 bg-brand-50 px-3 text-sm font-bold text-brand-700 transition hover:bg-brand-100 active:scale-95 ${className}`}
    >
      <Compass className="h-[1.15rem] w-[1.15rem]" />
      <span className="hidden sm:inline">Tour</span>
    </button>
  );
}
