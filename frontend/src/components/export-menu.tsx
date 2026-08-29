import { useEffect, useRef, useState } from "react";
import { Download, FileSpreadsheet, FileText, Loader2, Table2 } from "lucide-react";
import { ApiError, exportUrl } from "../lib/api";
import { toast } from "../lib/toast-store";

type Format = "xlsx" | "pdf" | "csv";

const FORMATS: { key: Format; label: string; hint: string; icon: typeof FileText }[] = [
  { key: "xlsx", label: "Excel", hint: "Angka siap dijumlah", icon: FileSpreadsheet },
  { key: "pdf", label: "PDF", hint: "Siap dicetak", icon: FileText },
  { key: "csv", label: "CSV", hint: "Untuk alat lain", icon: Table2 },
];

/**
 * Downloads a report in the format the user picks, carrying whatever filters
 * the screen currently has applied — so "export" means "export what I am
 * looking at".
 *
 * The file is fetched rather than linked because the session cookie is on a
 * different origin in development, and a failure needs to surface as a message
 * rather than a browser tab showing raw JSON.
 */
export function ExportMenu({
  report,
  params = {},
  label = "Export",
}: {
  report: string;
  params?: Record<string, string | number | boolean | undefined>;
  label?: string;
}) {
  const [open, setOpen] = useState(false);
  const [busy, setBusy] = useState<Format | null>(null);
  const box = useRef<HTMLDivElement>(null);

  // Close on an outside click or Escape, the way a menu is expected to behave.
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

  async function download(format: Format) {
    setBusy(format);
    try {
      const res = await fetch(exportUrl(report, { ...params, format }), {
        credentials: "include",
      });
      if (!res.ok) {
        const text = await res.text();
        let message = "Gagal mengunduh berkas.";
        try {
          message = JSON.parse(text)?.error ?? message;
        } catch {
          // A non-JSON body means the failure came from outside the API.
        }
        throw new ApiError(message, res.status);
      }

      const blob = await res.blob();
      // The server names the file; falling back keeps the extension right if
      // a proxy ever strips the header.
      const disposition = res.headers.get("Content-Disposition") ?? "";
      const named = /filename="([^"]+)"/.exec(disposition)?.[1];

      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = named ?? `${report}.${format}`;
      document.body.appendChild(link);
      link.click();
      link.remove();
      URL.revokeObjectURL(url);

      setOpen(false);
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Gagal mengunduh berkas.");
    } finally {
      setBusy(null);
    }
  }

  return (
    <div ref={box} className="relative shrink-0">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        aria-haspopup="menu"
        aria-expanded={open}
        className="flex min-h-11 items-center gap-2 rounded-xl border border-brand-100 bg-white px-3 text-sm font-bold text-ink transition hover:border-brand-200 hover:bg-brand-50"
      >
        {busy ? <Loader2 className="h-4 w-4 animate-spin" /> : <Download className="h-4 w-4" />}
        {label}
      </button>

      {open && (
        <div
          role="menu"
          className="absolute right-0 z-30 mt-1.5 w-56 overflow-hidden rounded-2xl border border-brand-100 bg-white p-1.5 shadow-xl"
        >
          {FORMATS.map((f) => (
            <button
              key={f.key}
              type="button"
              role="menuitem"
              disabled={busy !== null}
              onClick={() => download(f.key)}
              className="flex w-full items-center gap-3 rounded-xl px-2.5 py-2 text-left transition hover:bg-brand-50 disabled:opacity-50"
            >
              <span className="grid h-8 w-8 shrink-0 place-items-center rounded-lg bg-brand-50 text-brand-600">
                {busy === f.key ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <f.icon className="h-4 w-4" />
                )}
              </span>
              <span className="min-w-0">
                <span className="block text-sm font-bold text-ink">{f.label}</span>
                <span className="block truncate text-xs text-muted">{f.hint}</span>
              </span>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
