import { useEffect, useState, type FormEvent } from "react";
import { useSearchParams } from "react-router-dom";
import { Check, Gift, Handshake, Loader2, Search, UserPlus } from "lucide-react";
import { api, ApiError } from "../lib/api";
import { formatDate, today } from "../lib/format";
import { useApi } from "../lib/use-api";
import { PageHeader, Tabs } from "../components/ui";
import type { Member } from "../lib/types";

type Kind = "tyfcb" | "visitor" | "one-to-one";

export function SubmitPage() {
  const [params, setParams] = useSearchParams();
  const raw = params.get("type");
  const initial: Kind = raw === "visitor" || raw === "one-to-one" ? raw : "tyfcb";
  const [kind, setKind] = useState<Kind>(initial);

  function changeKind(next: Kind) {
    setKind(next);
    setParams({ type: next }, { replace: true });
  }

  return (
    <div className="space-y-5 lg:space-y-6">
      <PageHeader
        title="Catat Kontribusi"
        description="Semua submission berstatus pending sampai diverifikasi oleh Growth Coordinator."
      />

      <Tabs
        tabs={[
          { key: "tyfcb" as Kind, label: "TYFCB" },
          { key: "visitor" as Kind, label: "Visitor" },
          { key: "one-to-one" as Kind, label: "1-2-1" },
        ]}
        active={kind}
        onChange={changeKind}
      />

      {kind === "tyfcb" && <TyfcbForm />}
      {kind === "visitor" && <VisitorForm />}
      {kind === "one-to-one" && <OneToOneForm />}
    </div>
  );
}

function Result({ message, tone }: { message: string; tone: "ok" | "error" }) {
  return (
    <p
      role="alert"
      className={`mt-3 rounded-xl px-3 py-2 text-sm font-semibold ${
        tone === "ok" ? "bg-emerald-50 text-emerald-700" : "bg-red-50 text-red-700"
      }`}
    >
      {message}
    </p>
  );
}

function TyfcbForm() {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<Member[]>([]);
  const [buyer, setBuyer] = useState<Member | null>(null);
  const [nilai, setNilai] = useState("");
  const [tanggal, setTanggal] = useState(today());
  const [submitting, setSubmitting] = useState(false);
  const [result, setResult] = useState<{ message: string; tone: "ok" | "error" } | null>(null);

  // Whether the list shows is derived, not stored, so no effect has to clear
  // it — that would cost an extra render on every keystroke.
  const canSearch = !buyer && query.trim().length >= 3;
  const visibleResults = canSearch ? results : [];

  // Debounced so typing a name does not fire a request per keystroke.
  useEffect(() => {
    if (!canSearch) return;

    const timer = window.setTimeout(() => {
      api.members
        .search(query)
        .then(setResults)
        .catch(() => setResults([]));
    }, 250);

    return () => window.clearTimeout(timer);
  }, [query, canSearch]);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setResult(null);

    if (!buyer) {
      setResult({ message: "Pilih pemberi bisnis terlebih dahulu.", tone: "error" });
      return;
    }

    const amount = Number(nilai);
    if (!Number.isFinite(amount) || amount <= 0) {
      setResult({ message: "Nilai transaksi tidak valid.", tone: "error" });
      return;
    }

    setSubmitting(true);
    try {
      const entry = await api.tyfcb.submit(buyer.id, amount, tanggal);
      setResult({
        message: `Tersimpan. Skor sementara ${entry.computed_score} poin (pasangan ke-${entry.pair_ordinal}), menunggu verifikasi.`,
        tone: "ok",
      });
      setBuyer(null);
      setQuery("");
      setNilai("");
    } catch (err) {
      setResult({
        message: err instanceof ApiError ? err.message : "Gagal menyimpan.",
        tone: "error",
      });
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="card">
      <h2 className="flex items-center gap-2 text-base font-black text-brand-700">
        <Gift className="h-[1.1rem] w-[1.1rem]" />
        TYFCB
      </h2>
      <p className="mt-1.5 text-sm leading-relaxed text-muted">
        Catat bisnis yang kamu terima dari sesama member. Skor TYFCB masuk ke akun{" "}
        <strong className="text-ink">pembeli</strong> yang memberimu bisnis.
      </p>

      <div className="mt-4">
        <label className="section-label mb-1.5 block">Pemberi Bisnis (Pembeli) *</label>

        {buyer ? (
          <div className="flex items-center justify-between gap-3 rounded-xl border border-brand-600 bg-brand-50 px-3 py-2.5">
            <div className="min-w-0">
              <p className="truncate text-sm font-bold text-ink">{buyer.full_name}</p>
              <p className="truncate text-xs text-muted">{buyer.nama_tim ?? "—"}</p>
            </div>
            <button
              type="button"
              onClick={() => setBuyer(null)}
              className="shrink-0 text-xs font-bold text-brand-600 underline"
            >
              Ganti
            </button>
          </div>
        ) : (
          <>
            <span className="relative block">
              <Search
                aria-hidden
                className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted"
              />
              <input
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder="Ketik min. 3 huruf nama member"
                className="field pl-9"
              />
            </span>

            {visibleResults.length > 0 && (
              <ul className="mt-2 max-h-56 overflow-y-auto rounded-xl border border-brand-100">
                {visibleResults.map((member) => (
                  <li key={member.id}>
                    <button
                      type="button"
                      onClick={() => {
                        setBuyer(member);
                        setResults([]);
                      }}
                      className="flex w-full flex-col items-start px-3 py-2.5 text-left transition hover:bg-brand-50"
                    >
                      <span className="text-sm font-bold text-ink">{member.full_name}</span>
                      <span className="text-xs text-muted">{member.nama_tim ?? "—"}</span>
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </>
        )}
      </div>

      <label className="mt-4 block">
        <span className="section-label mb-1.5 block">Nilai transaksi *</span>
        <input
          type="number"
          min="1"
          required
          inputMode="numeric"
          value={nilai}
          onChange={(e) => setNilai(e.target.value)}
          placeholder="IDR"
          className="field num"
        />
      </label>

      <label className="mt-4 block">
        <span className="section-label mb-1.5 block">Tanggal transaksi *</span>
        <input
          type="date"
          required
          value={tanggal}
          onChange={(e) => setTanggal(e.target.value)}
          className="field"
        />
      </label>

      {result && <Result {...result} />}

      <button type="submit" disabled={submitting} className="btn-primary mt-5 w-full">
        {submitting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Check className="h-4 w-4" />}
        {submitting ? "Menyimpan…" : "Simpan TYFCB"}
      </button>
    </form>
  );
}

function VisitorForm() {
  const [nama, setNama] = useState("");
  const [kontak, setKontak] = useState("");
  const [tanggal, setTanggal] = useState(today());
  const [submitting, setSubmitting] = useState(false);
  const [result, setResult] = useState<{ message: string; tone: "ok" | "error" } | null>(null);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setResult(null);
    setSubmitting(true);

    try {
      await api.visitors.register(nama, kontak, tanggal);
      setResult({ message: "Visitor terdaftar. Poin menyusul setelah kehadiran dikonfirmasi.", tone: "ok" });
      setNama("");
      setKontak("");
    } catch (err) {
      setResult({
        message: err instanceof ApiError ? err.message : "Gagal menyimpan.",
        tone: "error",
      });
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="card">
      <h2 className="flex items-center gap-2 text-base font-black text-brand-700">
        <UserPlus className="h-[1.1rem] w-[1.1rem]" />
        Visitor
      </h2>
      <p className="mt-1.5 text-sm leading-relaxed text-muted">
        Daftarkan tamu yang kamu undang. Poin bertambah bertingkat saat tamu hadir (20), hadir penuh
        (50), dan berkonversi jadi member (+100).
      </p>

      <label className="mt-4 block">
        <span className="section-label mb-1.5 block">Nama tamu *</span>
        <input required value={nama} onChange={(e) => setNama(e.target.value)} className="field" />
      </label>

      <label className="mt-4 block">
        <span className="section-label mb-1.5 block">Kontak *</span>
        <input
          required
          value={kontak}
          onChange={(e) => setKontak(e.target.value)}
          placeholder="08xxxxxxxxxx"
          className="field num"
        />
        <span className="mt-1 block text-xs text-muted">
          Satu kontak hanya bisa didaftarkan sekali per season.
        </span>
      </label>

      <label className="mt-4 block">
        <span className="section-label mb-1.5 block">Tanggal undangan *</span>
        <input
          type="date"
          required
          value={tanggal}
          onChange={(e) => setTanggal(e.target.value)}
          className="field"
        />
      </label>

      {result && <Result {...result} />}

      <button type="submit" disabled={submitting} className="btn-primary mt-5 w-full">
        {submitting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Check className="h-4 w-4" />}
        {submitting ? "Menyimpan…" : "Daftarkan Visitor"}
      </button>
    </form>
  );
}

/**
 * Recording a 1-2-1 earns nothing on its own. During a 1-2-1 Payoff week, a
 * pair that both met and closed business is worth +30 each, settled by the
 * weekly pass — so the log is what makes that bonus provable.
 */
function OneToOneForm() {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<Member[]>([]);
  const [partner, setPartner] = useState<Member | null>(null);
  const [tanggal, setTanggal] = useState(today());
  const [catatan, setCatatan] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [result, setResult] = useState<{ message: string; tone: "ok" | "error" } | null>(null);
  const [reloadKey, setReloadKey] = useState(0);

  const canSearch = !partner && query.trim().length >= 3;
  const visibleResults = canSearch ? results : [];

  const { data: logs } = useApi(() => api.oneToOne.list(20), [reloadKey]);

  useEffect(() => {
    if (!canSearch) return;
    const timer = window.setTimeout(() => {
      api.members.search(query).then(setResults).catch(() => setResults([]));
    }, 250);
    return () => window.clearTimeout(timer);
  }, [query, canSearch]);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setResult(null);

    if (!partner) {
      setResult({ message: "Pilih lawan bicara terlebih dahulu.", tone: "error" });
      return;
    }

    setSubmitting(true);
    try {
      await api.oneToOne.log(partner.id, tanggal, catatan || null);
      setResult({ message: "1-2-1 tercatat.", tone: "ok" });
      setPartner(null);
      setQuery("");
      setCatatan("");
      setReloadKey((k) => k + 1);
    } catch (err) {
      setResult({
        message: err instanceof ApiError ? err.message : "Gagal menyimpan.",
        tone: "error",
      });
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="space-y-4">
      <form onSubmit={handleSubmit} className="card">
        <h2 className="flex items-center gap-2 text-base font-black text-brand-700">
          <Handshake className="h-[1.1rem] w-[1.1rem]" />
          1-2-1
        </h2>
        <p className="mt-1.5 text-sm leading-relaxed text-muted">
          Catat pertemuan empat mata dengan sesama member. Tidak berpoin sendiri — tapi di minggu
          1-2-1 Payoff, pasangan yang bertemu <em>dan</em> closing di minggu sama dapat +30 masing-masing.
        </p>

        <div className="mt-4">
          <label className="section-label mb-1.5 block">Lawan bicara *</label>

          {partner ? (
            <div className="flex items-center justify-between gap-3 rounded-xl border border-brand-600 bg-brand-50 px-3 py-2.5">
              <div className="min-w-0">
                <p className="truncate text-sm font-bold text-ink">{partner.full_name}</p>
                <p className="truncate text-xs text-muted">{partner.nama_tim ?? "—"}</p>
              </div>
              <button
                type="button"
                onClick={() => setPartner(null)}
                className="shrink-0 text-xs font-bold text-brand-600 underline"
              >
                Ganti
              </button>
            </div>
          ) : (
            <>
              <span className="relative block">
                <Search
                  aria-hidden
                  className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted"
                />
                <input
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  placeholder="Ketik min. 3 huruf nama member"
                  className="field pl-9"
                />
              </span>

              {visibleResults.length > 0 && (
                <ul className="mt-2 max-h-56 overflow-y-auto rounded-xl border border-brand-100">
                  {visibleResults.map((member) => (
                    <li key={member.id}>
                      <button
                        type="button"
                        onClick={() => {
                          setPartner(member);
                          setResults([]);
                        }}
                        className="flex w-full flex-col items-start px-3 py-2.5 text-left transition hover:bg-brand-50"
                      >
                        <span className="text-sm font-bold text-ink">{member.full_name}</span>
                        <span className="text-xs text-muted">{member.nama_tim ?? "—"}</span>
                      </button>
                    </li>
                  ))}
                </ul>
              )}
            </>
          )}
        </div>

        <label className="mt-4 block">
          <span className="section-label mb-1.5 block">Tanggal pertemuan *</span>
          <input
            type="date"
            required
            value={tanggal}
            onChange={(e) => setTanggal(e.target.value)}
            className="field"
          />
        </label>

        <label className="mt-4 block">
          <span className="section-label mb-1.5 block">Catatan</span>
          <textarea
            rows={2}
            value={catatan}
            onChange={(e) => setCatatan(e.target.value)}
            placeholder="Apa yang dibahas?"
            className="field resize-y"
          />
        </label>

        {result && <Result {...result} />}

        <button type="submit" disabled={submitting} className="btn-primary mt-5 w-full">
          {submitting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Check className="h-4 w-4" />}
          {submitting ? "Menyimpan…" : "Catat 1-2-1"}
        </button>
      </form>

      {logs && logs.length > 0 && (
        <section>
          <h3 className="section-label mb-2">Pertemuan terakhir</h3>
          <ul className="space-y-2">
            {logs.map((log) => (
              <li key={log.id} className="card flex items-center justify-between gap-3 p-3.5">
                <div className="min-w-0">
                  <p className="truncate text-sm font-bold text-ink">
                    {log.member_a_name} · {log.member_b_name}
                  </p>
                  {log.catatan && (
                    <p className="truncate text-xs text-muted">{log.catatan}</p>
                  )}
                </div>
                <span className="num shrink-0 text-xs text-muted">{formatDate(log.tanggal)}</span>
              </li>
            ))}
          </ul>
        </section>
      )}
    </div>
  );
}
