import { Component, type ErrorInfo, type ReactNode } from "react";
import { AlertTriangle, RotateCw } from "lucide-react";

type Props = { children: ReactNode };
type State = { error: Error | null };

/**
 * Without this, one render error blanks the page — React unmounts the whole
 * tree and the user sees white. A class component is required: hooks have no
 * equivalent of componentDidCatch.
 */
export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null };

  static getDerivedStateFromError(error: Error): State {
    return { error };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    // Kept in the console rather than sent anywhere: there is no error
    // reporting service wired up, and inventing one would hide the failure.
    console.error("render failed", error, info.componentStack);
  }

  render() {
    if (!this.state.error) return this.props.children;

    return (
      <main className="flex min-h-dvh items-center justify-center px-4">
        <div className="card w-full max-w-sm p-6 text-center">
          <span className="mx-auto grid h-12 w-12 place-items-center rounded-2xl bg-red-50 text-red-600">
            <AlertTriangle className="h-6 w-6" />
          </span>

          <h1 className="mt-4 text-lg font-black text-ink">Ada yang salah</h1>
          <p className="mt-1.5 text-sm leading-relaxed text-muted">
            Halaman ini gagal dimuat. Coba muat ulang — kalau terus terjadi, laporkan ke Growth
            Coordinator.
          </p>

          <button
            type="button"
            onClick={() => window.location.reload()}
            className="btn-primary mt-5 w-full"
          >
            <RotateCw className="h-4 w-4" />
            Muat ulang
          </button>
        </div>
      </main>
    );
  }
}
