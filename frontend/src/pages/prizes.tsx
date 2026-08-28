import { useState, type FormEvent } from "react";
import { Gift, Loader2, Ticket, Trophy } from "lucide-react";
import { api, ApiError } from "../lib/api";
import { useApi } from "../lib/use-api";
import { formatCurrency } from "../lib/format";
import { Badge, EmptyState, ErrorNote, PageHeader, Spinner, Tabs } from "../components/ui";

type Tab = "pool" | "tickets" | "donate";

export function PrizesPage() {
  const [tab, setTab] = useState<Tab>("pool");

  return (
    <div className="space-y-5">
      <PageHeader
        title="Prize Pool"
        description="Hadiah dialokasikan dua lapis: sebagian untuk pemenang kategori leaderboard, sisanya diundi. Tiket undian dihitung dari poin, visitor hadir, dan TYFCB ke pasangan baru."
      />

      <Tabs
        tabs={[
          { key: "pool" as Tab, label: "Hadiah" },
          { key: "tickets" as Tab, label: "Tiket Undian" },
          { key: "donate" as Tab, label: "Donasi" },
        ]}
        active={tab}
        onChange={setTab}
      />

      {tab === "pool" && <PrizeList />}
      {tab === "tickets" && <TicketList />}
      {tab === "donate" && <DonateForm onDone={() => setTab("pool")} />}
    </div>
  );
}

function PrizeList() {
  const { data, error, loading, reload } = useApi(() => api.prizes.list());

  if (loading) return <Spinner />;
  if (error) return <ErrorNote message={error} onRetry={reload} />;

  const prizes = (data ?? []).filter((p) => p.status !== "rejected");
  if (prizes.length === 0) return <EmptyState message="Belum ada hadiah di pool." />;

  return (
    <ul className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
      {prizes.map((prize) => (
        <li key={prize.id} className="card p-4">
          <div className="flex items-start gap-3">
            <span className="grid h-10 w-10 shrink-0 place-items-center rounded-xl bg-brand-50 text-brand-600">
              {prize.alokasi === "undian" ? <Ticket className="h-5 w-5" /> : <Trophy className="h-5 w-5" />}
            </span>
            <div className="min-w-0 flex-1">
              <p className="truncate text-base font-black text-ink">{prize.nama_hadiah}</p>
              {prize.deskripsi && (
                <p className="mt-0.5 text-xs leading-relaxed text-muted">{prize.deskripsi}</p>
              )}
              <div className="mt-2 flex flex-wrap items-center gap-x-2 gap-y-1 text-xs">
                <Badge value={prize.status} />
                <span className="text-muted">
                  {prize.alokasi === "undian" ? "Diundi" : `Kategori${prize.kategori_target ? `: ${prize.kategori_target}` : ""}`}
                </span>
                {prize.nilai_estimasi != null && (
                  <span className="num text-muted">· {formatCurrency(prize.nilai_estimasi)}</span>
                )}
              </div>
              {prize.donatur_nama && (
                <p className="mt-1 text-xs text-muted">Donasi dari {prize.donatur_nama}</p>
              )}
              {prize.pemenang_nama && (
                <p className="num mt-1 text-xs font-bold text-emerald-700">
                  Pemenang: {prize.pemenang_nama}
                </p>
              )}
            </div>
          </div>
        </li>
      ))}
    </ul>
  );
}

function TicketList() {
  const { data, error, loading, reload } = useApi(() => api.raffle.tickets());

  if (loading) return <Spinner />;
  if (error) return <ErrorNote message={error} onRetry={reload} />;

  const tickets = [...(data ?? [])].sort((a, b) => b.tickets - a.tickets);
  if (tickets.length === 0) {
    return <EmptyState message="Tiket belum diterbitkan. Admin menerbitkannya dari panel Event." />;
  }

  return (
    <ol className="space-y-1.5">
      {tickets.map((row, index) => (
        <li
          key={row.member_id}
          className="card grid grid-cols-[1.75rem_1fr_auto] items-center gap-2.5 px-3 py-2.5"
        >
          <span className="num text-center text-sm font-black text-muted">{index + 1}</span>
          <div className="min-w-0">
            <p className="truncate text-sm font-bold text-ink">{row.full_name}</p>
            <p className="truncate text-xs text-muted">{row.nama_tim ?? "—"}</p>
          </div>
          <span className="num shrink-0 text-sm font-black text-brand-600">
            {row.tickets} tiket
          </span>
        </li>
      ))}
    </ol>
  );
}

function DonateForm({ onDone }: { onDone: () => void }) {
  const [form, setForm] = useState({
    nama_hadiah: "",
    deskripsi: "",
    nilai_estimasi: "",
    alokasi: "undian",
  });
  const [busy, setBusy] = useState(false);
  const [result, setResult] = useState<{ message: string; tone: "ok" | "error" } | null>(null);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setResult(null);
    setBusy(true);

    try {
      await api.prizes.donate({
        nama_hadiah: form.nama_hadiah,
        deskripsi: form.deskripsi || null,
        nilai_estimasi: form.nilai_estimasi ? Number(form.nilai_estimasi) : null,
        alokasi: form.alokasi,
        kategori_target: null,
      });
      setResult({ message: "Terima kasih. Donasi menunggu persetujuan admin.", tone: "ok" });
      setForm({ nama_hadiah: "", deskripsi: "", nilai_estimasi: "", alokasi: "undian" });
      setTimeout(onDone, 1200);
    } catch (err) {
      setResult({
        message: err instanceof ApiError ? err.message : "Gagal menyimpan.",
        tone: "error",
      });
    } finally {
      setBusy(false);
    }
  }

  return (
    <form onSubmit={submit} className="card p-4 sm:p-5">
      <h2 className="flex items-center gap-2 text-base font-black text-brand-700">
        <Gift className="h-[1.1rem] w-[1.1rem]" />
        Donasi hadiah
      </h2>
      <p className="mt-1.5 text-sm leading-relaxed text-muted">
        Sumbangkan produk atau jasa ke prize pool. Donasi yang disetujui admin memberimu badge
        Patron.
      </p>

      <label className="mt-4 block">
        <span className="section-label mb-1.5 block">Nama hadiah *</span>
        <input
          required
          value={form.nama_hadiah}
          onChange={(e) => setForm({ ...form, nama_hadiah: e.target.value })}
          className="field"
        />
      </label>

      <label className="mt-3 block">
        <span className="section-label mb-1.5 block">Deskripsi</span>
        <textarea
          rows={3}
          value={form.deskripsi}
          onChange={(e) => setForm({ ...form, deskripsi: e.target.value })}
          className="field resize-y"
        />
      </label>

      <div className="mt-3 grid grid-cols-2 gap-3">
        <label className="block">
          <span className="section-label mb-1.5 block">Estimasi nilai</span>
          <input
            type="number"
            min="0"
            value={form.nilai_estimasi}
            onChange={(e) => setForm({ ...form, nilai_estimasi: e.target.value })}
            placeholder="IDR"
            className="field num"
          />
        </label>

        <label className="block">
          <span className="section-label mb-1.5 block">Alokasi</span>
          <select
            value={form.alokasi}
            onChange={(e) => setForm({ ...form, alokasi: e.target.value })}
            className="field"
          >
            <option value="undian">Diundi</option>
            <option value="kategori">Pemenang kategori</option>
          </select>
        </label>
      </div>

      {result && (
        <p
          role="alert"
          className={`mt-3 rounded-xl px-3 py-2 text-sm font-semibold ${
            result.tone === "ok" ? "bg-emerald-50 text-emerald-700" : "bg-red-50 text-red-700"
          }`}
        >
          {result.message}
        </p>
      )}

      <button type="submit" disabled={busy} className="btn-primary mt-5 w-full">
        {busy && <Loader2 className="h-4 w-4 animate-spin" />}
        Kirim donasi
      </button>
    </form>
  );
}
