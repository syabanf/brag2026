/**
 * Toast store, kept apart from the component so that file exports only a
 * component. Module scope means any handler can raise a toast without
 * threading a context through the tree.
 */
export type Toast = {
  id: number;
  message: string;
  tone: "ok" | "error";
};

let toasts: Toast[] = [];
let nextId = 1;
const listeners = new Set<() => void>();

function emit() {
  listeners.forEach((listener) => listener());
}

export function dismissToast(id: number) {
  toasts = toasts.filter((t) => t.id !== id);
  emit();
}

function push(message: string, tone: Toast["tone"]) {
  const id = nextId++;
  toasts = [...toasts, { id, message, tone }];
  emit();

  // Errors linger longer: they usually need reading, successes only glancing.
  window.setTimeout(() => dismissToast(id), tone === "error" ? 6000 : 3500);
}

export const toast = {
  ok: (message: string) => push(message, "ok"),
  error: (message: string) => push(message, "error"),
};

export function subscribeToasts(listener: () => void) {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

export const getToasts = () => toasts;
export const getServerToasts = () => [] as Toast[];
