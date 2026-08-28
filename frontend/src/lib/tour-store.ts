/**
 * Module-scoped store for the guided tour. Kept apart from the component so
 * that file exports only components, which is what React Fast Refresh needs.
 * Module state also outlives the shell remount that every navigation causes,
 * which is exactly what a tour that walks across routes requires.
 */
export type TourState = { active: boolean; index: number; muted: boolean };

export const INITIAL_TOUR: TourState = { active: false, index: 0, muted: false };

let state: TourState = INITIAL_TOUR;
const listeners = new Set<() => void>();

export function setTourState(next: TourState) {
  state = next;
  listeners.forEach((listener) => listener());
}

export function getTourState() {
  return state;
}

export function subscribeTour(listener: () => void) {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

export function startTour() {
  setTourState({ active: true, index: 0, muted: false });
}
