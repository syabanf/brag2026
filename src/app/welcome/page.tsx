import { isDemoAvailable } from "@/lib/demo";
import { WelcomeChoices } from "./welcome-client";

export const metadata = {
  title: "BRAG 2026 — Mulai"
};

export default function WelcomePage() {
  return (
    <main className="min-h-dvh bg-gradient-to-b from-brand-50 via-white to-white px-4 py-10 sm:px-6">
      <div className="mx-auto w-full max-w-md">
        <header className="mb-7 text-center">
          <h1 className="text-4xl font-black tracking-normal text-brand-600 sm:text-5xl">
            BRAG 2026
          </h1>
          <p className="mt-1.5 text-[0.68rem] font-bold uppercase tracking-[0.18em] text-brand-700">
            BNI Grow Annual Challenge
          </p>
          <p className="mx-auto mt-4 max-w-sm text-sm leading-relaxed text-muted">
            Platform gamifikasi kompetisi anggota. Pilih cara Anda ingin masuk.
          </p>
        </header>

        <WelcomeChoices demoAvailable={isDemoAvailable()} />

        <p className="mt-6 text-center text-xs leading-relaxed text-muted">
          Mode demo berjalan di database sementara dalam memori. Perubahan apa pun
          hilang saat server dimatikan.
        </p>
      </div>
    </main>
  );
}
