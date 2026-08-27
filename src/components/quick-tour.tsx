"use client";

import { useCallback, useEffect, useRef, useSyncExternalStore } from "react";
import { usePathname, useRouter } from "next/navigation";
import { ChevronLeft, ChevronRight, Volume2, VolumeX, X } from "lucide-react";
import { TOUR_STEPS } from "@/lib/tour-steps";

export const TOUR_EVENT = "brag:start-tour";

type TourState = { active: boolean; index: number; muted: boolean };

const INITIAL: TourState = { active: false, index: 0, muted: false };

// Module-scoped store rather than component state: the tour walks the user
// across routes, and AppShell remounts on every navigation. Module state
// outlives those remounts, and useSyncExternalStore keeps SSR rendering
// nothing so there is no hydration mismatch.
let tourState: TourState = INITIAL;
const listeners = new Set<() => void>();

function setTourState(next: TourState) {
  tourState = next;
  listeners.forEach((listener) => listener());
}

function subscribe(listener: () => void) {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

function getSnapshot() {
  return tourState;
}

function getServerSnapshot() {
  return INITIAL;
}

function speakFallback(text: string) {
  if (typeof window === "undefined" || !window.speechSynthesis) return;

  window.speechSynthesis.cancel();
  const utterance = new SpeechSynthesisUtterance(text);
  utterance.lang = "id-ID";
  utterance.rate = 1;

  const indonesian = window.speechSynthesis
    .getVoices()
    .find((voice) => voice.lang.toLowerCase().startsWith("id"));
  if (indonesian) utterance.voice = indonesian;

  window.speechSynthesis.speak(utterance);
}

export function QuickTour() {
  const router = useRouter();
  const pathname = usePathname();
  const state = useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot);
  const audioRef = useRef<HTMLAudioElement | null>(null);

  const step = TOUR_STEPS[state.index];
  const isLast = state.index === TOUR_STEPS.length - 1;

  useEffect(() => {
    const onStart = () => setTourState({ active: true, index: 0, muted: false });
    window.addEventListener(TOUR_EVENT, onStart);
    return () => window.removeEventListener(TOUR_EVENT, onStart);
  }, []);

  const stopAudio = useCallback(() => {
    audioRef.current?.pause();
    audioRef.current = null;
  }, []);

  // Narration follows the active step and is cancelled whenever it changes.
  // ElevenLabs is preferred; when it is unavailable (no key, or a voice the
  // plan does not cover) the browser's own speech synthesis takes over so the
  // tour is never silent.
  useEffect(() => {
    if (!state.active || state.muted || !step) return;

    let cancelled = false;
    let objectUrl: string | null = null;
    const audio = new Audio();
    audioRef.current = audio;

    async function narrate() {
      try {
        const res = await fetch(`/api/tour/voice?step=${step.id}`);
        const type = res.headers.get("content-type") ?? "";

        if (!cancelled && res.ok && type.startsWith("audio/")) {
          objectUrl = URL.createObjectURL(await res.blob());
          if (cancelled) return;
          audio.src = objectUrl;
          await audio.play();
          return;
        }
      } catch {
        // Network or autoplay failure — fall through to speech synthesis.
      }

      if (cancelled) return;
      speakFallback(step.voice);
    }

    void narrate();

    return () => {
      cancelled = true;
      audio.pause();
      if (objectUrl) URL.revokeObjectURL(objectUrl);
      if (typeof window !== "undefined") window.speechSynthesis?.cancel();
    };
  }, [state.active, state.muted, step]);

  useEffect(() => {
    if (!state.active || !step) return;
    if (pathname !== step.route) return;
    if (!step.selector) return;

    const timer = window.setTimeout(() => {
      const el = document.querySelector(step.selector as string);
      el?.scrollIntoView({ behavior: "smooth", block: "center" });
    }, 350);

    return () => window.clearTimeout(timer);
  }, [state.active, step, pathname]);

  const goTo = useCallback(
    (index: number) => {
      const next = TOUR_STEPS[index];
      if (!next) return;
      stopAudio();
      setTourState({ ...tourState, index });
      if (next.route !== pathname) router.push(next.route);
    },
    [pathname, router, stopAudio]
  );

  const close = useCallback(() => {
    stopAudio();
    setTourState(INITIAL);
  }, [stopAudio]);

  if (!state.active || !step) return null;

  return (
    <>
      <div
        aria-hidden
        className="fixed inset-0 z-40 bg-ink/40 backdrop-blur-[2px]"
        onClick={close}
      />

      <div
        role="dialog"
        aria-modal="true"
        aria-label={`Quick tour: ${step.title}`}
        className="fixed inset-x-0 bottom-0 z-50 px-3 pb-[calc(env(safe-area-inset-bottom)+0.75rem)] sm:px-4 lg:inset-x-auto lg:right-6 lg:bottom-6 lg:w-[26rem] lg:px-0 lg:pb-0"
      >
        <div className="rounded-3xl border border-brand-100 bg-white p-5 shadow-[0_18px_60px_rgba(80,0,18,0.28)]">
          <div className="mb-3 flex items-start justify-between gap-3">
            <div className="min-w-0">
              <p className="text-[0.68rem] font-bold uppercase tracking-[0.16em] text-brand-600">
                Quick Tour · {state.index + 1}/{TOUR_STEPS.length}
              </p>
              <h2 className="mt-1 text-lg font-black leading-tight text-ink">{step.title}</h2>
            </div>

            <div className="flex shrink-0 items-center gap-1">
              <button
                type="button"
                aria-label={state.muted ? "Nyalakan suara" : "Matikan suara"}
                aria-pressed={state.muted}
                onClick={() => {
                  stopAudio();
                  setTourState({ ...tourState, muted: !tourState.muted });
                }}
                className="flex h-10 w-10 items-center justify-center rounded-full text-muted transition hover:bg-brand-50 hover:text-brand-600"
              >
                {state.muted ? <VolumeX className="h-5 w-5" /> : <Volume2 className="h-5 w-5" />}
              </button>
              <button
                type="button"
                aria-label="Tutup tur"
                onClick={close}
                className="flex h-10 w-10 items-center justify-center rounded-full text-muted transition hover:bg-brand-50 hover:text-brand-600"
              >
                <X className="h-5 w-5" />
              </button>
            </div>
          </div>

          <p className="text-sm leading-relaxed text-muted">{step.body}</p>

          <div className="mt-4 flex items-center gap-1.5" aria-hidden>
            {TOUR_STEPS.map((s, i) => (
              <span
                key={s.id}
                className={`h-1.5 flex-1 rounded-full transition ${
                  i <= state.index ? "bg-brand-600" : "bg-brand-100"
                }`}
              />
            ))}
          </div>

          <div className="mt-4 flex items-center gap-2">
            <button
              type="button"
              onClick={() => goTo(state.index - 1)}
              disabled={state.index === 0}
              className="flex min-h-12 items-center gap-1 rounded-full border border-brand-100 px-4 text-sm font-semibold text-ink transition hover:bg-brand-50 disabled:opacity-40"
            >
              <ChevronLeft className="h-4 w-4" />
              Kembali
            </button>

            <button
              type="button"
              onClick={() => (isLast ? close() : goTo(state.index + 1))}
              className="flex min-h-12 flex-1 items-center justify-center gap-1 rounded-full bg-brand-600 px-4 text-sm font-bold text-white transition hover:bg-brand-700 active:scale-[0.98]"
            >
              {isLast ? "Selesai" : "Lanjut"}
              {!isLast && <ChevronRight className="h-4 w-4" />}
            </button>
          </div>
        </div>
      </div>
    </>
  );
}
