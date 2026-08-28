import { useCallback, useEffect, useRef, useState, useSyncExternalStore } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { ChevronLeft, ChevronRight, Compass, Volume2, VolumeX, X } from "lucide-react";
import { api } from "../lib/api";
import type { TourStep } from "../lib/types";
import {
  INITIAL_TOUR,
  getTourState,
  setTourState,
  startTour,
  subscribeTour,
} from "../lib/tour-store";

/** Falls back to the browser's own voice when ElevenLabs is unavailable. */
function speakFallback(text: string) {
  if (typeof window === "undefined" || !window.speechSynthesis) return;

  window.speechSynthesis.cancel();
  const utterance = new SpeechSynthesisUtterance(text);
  utterance.lang = "id-ID";

  const indonesian = window.speechSynthesis
    .getVoices()
    .find((voice) => voice.lang.toLowerCase().startsWith("id"));
  if (indonesian) utterance.voice = indonesian;

  window.speechSynthesis.speak(utterance);
}

export function TourButton({ className = "" }: { className?: string }) {
  return (
    <button
      type="button"
      onClick={startTour}
      aria-label="Mulai quick tour"
      className={`flex min-h-11 items-center gap-1.5 rounded-full border border-brand-100 bg-brand-50 px-3 text-sm font-bold text-brand-700 transition hover:bg-brand-100 active:scale-95 ${className}`}
    >
      <Compass className="h-[1.15rem] w-[1.15rem]" />
      <span className="hidden sm:inline">Tour</span>
    </button>
  );
}

export function QuickTour() {
  const navigate = useNavigate();
  const { pathname } = useLocation();
  const state = useSyncExternalStore(subscribeTour, getTourState, () => INITIAL_TOUR);
  const [steps, setSteps] = useState<TourStep[]>([]);
  const audioRef = useRef<HTMLAudioElement | null>(null);

  const step = steps[state.index];
  const isLast = state.index === steps.length - 1;

  useEffect(() => {
    if (!state.active || steps.length > 0) return;
    api.tour
      .steps()
      .then(setSteps)
      .catch(() => setSteps([]));
  }, [state.active, steps.length]);

  const stopAudio = useCallback(() => {
    audioRef.current?.pause();
    audioRef.current = null;
    window.speechSynthesis?.cancel();
  }, []);

  // Narration follows the active step. ElevenLabs is preferred; anything other
  // than audio means the browser voice takes over, so the tour is never silent.
  useEffect(() => {
    if (!state.active || state.muted || !step) return;

    let cancelled = false;
    let objectUrl: string | null = null;
    const audio = new Audio();
    audioRef.current = audio;

    async function narrate() {
      try {
        const blob = await api.tour.voice(step.id);
        if (cancelled) return;

        if (blob) {
          objectUrl = URL.createObjectURL(blob);
          audio.src = objectUrl;
          await audio.play();
          return;
        }
      } catch {
        // Network or autoplay failure — fall through to speech synthesis.
      }

      if (!cancelled) speakFallback(step.body);
    }

    void narrate();

    return () => {
      cancelled = true;
      audio.pause();
      if (objectUrl) URL.revokeObjectURL(objectUrl);
      window.speechSynthesis?.cancel();
    };
  }, [state.active, state.muted, step]);

  const goTo = useCallback(
    (index: number) => {
      const next = steps[index];
      if (!next) return;
      stopAudio();
      setTourState({ ...getTourState(), index });
      if (next.route !== pathname) navigate(next.route);
    },
    [steps, pathname, navigate, stopAudio],
  );

  const close = useCallback(() => {
    stopAudio();
    setTourState(INITIAL_TOUR);
  }, [stopAudio]);

  if (!state.active || !step) return null;

  return (
    <>
      <div aria-hidden className="fixed inset-0 z-40 bg-ink/40 backdrop-blur-[2px]" onClick={close} />

      <div
        role="dialog"
        aria-modal="true"
        aria-label={`Quick tour: ${step.title}`}
        className="fixed inset-x-0 bottom-0 z-50 px-3 pb-[calc(env(safe-area-inset-bottom)+0.75rem)] sm:px-4 lg:inset-x-auto lg:bottom-6 lg:right-6 lg:w-[26rem] lg:px-0 lg:pb-0"
      >
        <div className="rounded-3xl border border-brand-100 bg-white p-5 shadow-[0_18px_60px_rgba(80,0,18,0.28)]">
          <div className="mb-3 flex items-start justify-between gap-3">
            <div className="min-w-0">
              <p className="num text-[0.68rem] font-bold uppercase tracking-[0.16em] text-brand-600">
                Quick Tour · {state.index + 1}/{steps.length}
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
                  setTourState({ ...getTourState(), muted: !getTourState().muted });
                }}
                className="grid h-10 w-10 place-items-center rounded-full text-muted transition hover:bg-brand-50 hover:text-brand-600"
              >
                {state.muted ? <VolumeX className="h-5 w-5" /> : <Volume2 className="h-5 w-5" />}
              </button>
              <button
                type="button"
                aria-label="Tutup tur"
                onClick={close}
                className="grid h-10 w-10 place-items-center rounded-full text-muted transition hover:bg-brand-50 hover:text-brand-600"
              >
                <X className="h-5 w-5" />
              </button>
            </div>
          </div>

          <p className="text-sm leading-relaxed text-muted">{step.body}</p>

          <div className="mt-4 flex items-center gap-1.5" aria-hidden>
            {steps.map((s, i) => (
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
              className="btn-secondary px-4 disabled:opacity-40"
            >
              <ChevronLeft className="h-4 w-4" />
              Kembali
            </button>

            <button
              type="button"
              onClick={() => (isLast ? close() : goTo(state.index + 1))}
              className="btn-primary flex-1"
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
