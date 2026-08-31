import { useState, type FormEvent } from "react";
import { CalendarDays, Dices, Loader2, Network, Play, Ticket, Trash2, Trophy } from "lucide-react";
import { api, ApiError } from "../lib/api";
import { toast } from "../lib/toast-store";
import { useApi } from "../lib/use-api";
import { formatDate, today } from "../lib/format";
import { EmptyState, ErrorNote, PageHeader, Spinner, Tabs } from "../components/ui";
import { ExportMenu } from "../components/export-menu";
import type { PassResult } from "../lib/types";

type Tab = "schedule" | "passes" | "spheres" | "prizes";

export function AdminEventsPage() {
  const [tab, setTab] = useState<Tab>("schedule");

  return (
    <div className="space-y-5 lg:space-y-6">
      <PageHeader
        eyebrow="Admin Area"
        title="Event & Bonus"
        description="Jadwalkan event mingguan yang mengubah pengali poin, lalu jalankan pass berkala untuk menyelesaikan bonus yang bergantung pada satu hari atau minggu penuh."
      />

      <Tabs
        tabs={[
          { key: "schedule" as Tab, label: "Jadwal Event" },
          { key: "passes" as Tab, label: "Pass Berkala" },
          { key: "spheres" as Tab, label: "Contact Sphere" },
          { key: "prizes" as Tab, label: "Prize Pool" },
        ]}
        active={tab}
        onChange={setTab}
      />

      {tab === "schedule" && <EventSchedule />}
      {tab === "passes" && <PassRunner />}
      {tab === "spheres" && <SphereManager />}
      {tab === "prizes" && <PrizeAdmin />}
    </div>
  );
}

function EventSchedule() {
  const { data, error, loading, reload } = useApi(() => api.admin.events.list());
  const { data: bank } = useApi(() => api.events.bank());
  const [busy, setBusy] = useState(false);
  const [form, setForm] = useState({
    minggu_ke: "1",
    event_code: "",
    tanggal_mulai: today(),
    tanggal_selesai: today(),
  });

  async function schedule(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    try {
      await api.admin.events.schedule({
        minggu_ke: Number(form.minggu_ke),
        event_code: form.event_code,
        target_classification_id: null,
        tanggal_mulai: form.tanggal_mulai,
        tanggal_selesai: form.tanggal_selesai,
      });
      reload();
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Gagal menjadwalkan.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="space-y-4">
      <form onSubmit={schedule} className="card space-y-3 p-4">
        <h2 className="flex items-center gap-2 text-base font-black text-ink">
          <CalendarDays className="h-[1.1rem] w-[1.1rem] text-brand-600" />
          Jadwalkan event
        </h2>
        <p className="text-xs leading-relaxed text-muted">
          Satu event per minggu. Menjadwalkan ulang minggu yang sama akan menggantikan event
          sebelumnya.
        </p>

        <div className="grid grid-cols-2 gap-3">
          <label className="block">
            <span className="section-label mb-1.5 block">Minggu ke</span>
            <input
              type="number"
              min="1"
              max="12"
              required
              value={form.minggu_ke}
              onChange={(e) => setForm({ ...form, minggu_ke: e.target.value })}
              className="field num"
            />
          </label>

          <label className="block">
            <span className="section-label mb-1.5 block">Event</span>
            <select
              required
              value={form.event_code}
              onChange={(e) => setForm({ ...form, event_code: e.target.value })}
              className="field"
            >
              <option value="">— Pilih event</option>
              {(bank ?? []).map((entry) => (
                <option key={entry.code} value={entry.code}>
                  {entry.nama}
                </option>
              ))}
            </select>
          </label>
        </div>

        {form.event_code && (
          <p className="rounded-xl bg-brand-50 px-3 py-2 text-xs font-semibold text-brand-700">
            {bank?.find((b) => b.code === form.event_code)?.mekanik}
          </p>
        )}

        <div className="grid grid-cols-2 gap-3">
          <label className="block">
            <span className="section-label mb-1.5 block">Mulai</span>
            <input
              type="date"
              required
              value={form.tanggal_mulai}
              onChange={(e) => setForm({ ...form, tanggal_mulai: e.target.value })}
              className="field"
            />
          </label>
          <label className="block">
            <span className="section-label mb-1.5 block">Selesai</span>
            <input
              type="date"
              required
              value={form.tanggal_selesai}
              onChange={(e) => setForm({ ...form, tanggal_selesai: e.target.value })}
              className="field"
            />
          </label>
        </div>

        <button type="submit" disabled={busy} className="btn-primary w-full">
          {busy && <Loader2 className="h-4 w-4 animate-spin" />}
          Jadwalkan
        </button>
      </form>

      {loading && <Spinner />}
      {error && <ErrorNote message={error} onRetry={reload} />}
      {data && data.length === 0 && <EmptyState message="Belum ada event terjadwal." />}

      <ul className="space-y-2">
        {data?.map((event) => (
          <li key={event.id} className="card flex items-center gap-3 p-3.5">
            <span className="num grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-brand-50 text-xs font-black text-brand-700">
              W{event.minggu_ke}
            </span>
            <div className="min-w-0 flex-1">
              <p className="truncate text-sm font-black text-ink">{event.nama}</p>
              <p className="truncate text-xs text-muted">{event.mekanik}</p>
              <p className="num text-[0.68rem] text-muted">
                {formatDate(event.tanggal_mulai)} — {formatDate(event.tanggal_selesai)}
              </p>
            </div>
            <button
              type="button"
              aria-label={`Hapus ${event.nama}`}
              onClick={async () => {
                if (!confirm(`Hapus event ${event.nama}?`)) return;
                await api.admin.events.remove(event.id);
                reload();
              }}
              className="grid h-10 w-10 shrink-0 place-items-center rounded-lg border border-red-100 text-red-500 transition hover:bg-red-50"
            >
              <Trash2 className="h-3.5 w-3.5" />
            </button>
          </li>
        ))}
      </ul>
    </div>
  );
}

function PassRunner() {
  const [day, setDay] = useState(today());
  const [busy, setBusy] = useState<string | null>(null);
  const [result, setResult] = useState<PassResult | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function run(kind: "weekly" | "daily") {
    setBusy(kind);
    setError(null);
    setResult(null);
    try {
      setResult(await api.admin.passes[kind](day));
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Gagal menjalankan pass.");
    } finally {
      setBusy(null);
    }
  }

  return (
    <div className="space-y-4">
      <section className="card">
        <h2 className="text-base font-black text-ink">Jalankan pass</h2>
        <p className="mt-1.5 text-sm leading-relaxed text-muted">
          Pass menyelesaikan bonus yang bergantung pada periode: Full Roster dan Streak Week per
          minggu, High Roller Day per hari. Setiap pass idempoten per periode — menjalankannya dua
          kali tidak membayar dua kali.
        </p>

        <label className="mt-4 block">
          <span className="section-label mb-1.5 block">Tanggal acuan</span>
          <input
            type="date"
            value={day}
            onChange={(e) => setDay(e.target.value)}
            className="field"
          />
        </label>

        <div className="mt-4 grid grid-cols-2 gap-2">
          <button
            type="button"
            disabled={busy !== null}
            onClick={() => run("weekly")}
            className="btn-primary"
          >
            {busy === "weekly" ? <Loader2 className="h-4 w-4 animate-spin" /> : <Play className="h-4 w-4" />}
            Mingguan
          </button>
          <button
            type="button"
            disabled={busy !== null}
            onClick={() => run("daily")}
            className="btn-secondary"
          >
            {busy === "daily" ? <Loader2 className="h-4 w-4 animate-spin" /> : <Play className="h-4 w-4" />}
            Harian
          </button>
        </div>
      </section>

      {error && <ErrorNote message={error} />}

      {result && (
        <section className="card">
          <h3 className="section-label mb-2">Hasil · {result.period}</h3>
          <dl className="space-y-1.5 text-sm">
            <Row label="Poin ditambahkan" value={`${result.points_added} pts`} />
            <Row
              label="Full Roster"
              value={
                result.full_roster_teams.length > 0
                  ? result.full_roster_teams.join(", ")
                  : "tidak ada tim yang memenuhi"
              }
            />
            <Row label="Bonus streak" value={`${result.streak_awards} anggota`} />
            {result.high_roller_member && (
              <Row label="High Roller" value={result.high_roller_member} />
            )}
          </dl>

          {result.skipped && result.skipped.length > 0 && (
            <p className="mt-3 rounded-xl bg-amber-50 px-3 py-2 text-xs font-semibold text-amber-800">
              Dilewati: {result.skipped.join("; ")}
            </p>
          )}
        </section>
      )}
    </div>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-start justify-between gap-3">
      <dt className="text-muted">{label}</dt>
      <dd className="num text-right font-bold text-ink">{value}</dd>
    </div>
  );
}

function PrizeAdmin() {
  const { data, error, loading, reload } = useApi(() => api.admin.prizes.list());
  const [busy, setBusy] = useState<string | null>(null);
  const [issuing, setIssuing] = useState(false);

  async function setStatus(id: string, status: string) {
    setBusy(id);
    try {
      await api.admin.prizes.setStatus(id, status);
      reload();
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Gagal memperbarui.");
    } finally {
      setBusy(null);
    }
  }

  // Drawing is announced in the room and cannot be taken back, so it asks
  // first. The server refuses a second draw regardless.
  async function draw(id: string, nama: string) {
    if (!window.confirm(`Undi pemenang untuk "${nama}"? Hasil undian tidak bisa diulang.`)) return;

    setBusy(id);
    try {
      const won = await api.admin.raffle.draw(id);
      toast.ok(`Pemenang: ${won.pemenang_nama ?? "—"}`);
      reload();
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Gagal mengundi.");
    } finally {
      setBusy(null);
    }
  }

  return (
    <div className="space-y-4">
      <section className="card">
        <h2 className="flex items-center gap-2 text-base font-black text-ink">
          <Ticket className="h-[1.1rem] w-[1.1rem] text-brand-600" />
          Terbitkan tiket undian
        </h2>
        <p className="mt-1.5 text-sm leading-relaxed text-muted">
          Menghitung ulang hak tiket setiap anggota dari poin, visitor hadir, dan TYFCB ke pasangan
          baru. Menulis ulang, bukan menambah — aman dijalankan berkali-kali.
        </p>

        <div className="mt-3">
          <ExportMenu report="prizes" label="Export hadiah & tiket" />
        </div>
        <button
          type="button"
          disabled={issuing}
          onClick={async () => {
            setIssuing(true);
            try {
              const rows = await api.admin.raffle.issue();
              toast.ok(`Tiket diterbitkan untuk ${rows.length} anggota.`);
            } catch (err) {
              toast.error(err instanceof ApiError ? err.message : "Gagal menerbitkan tiket.");
            } finally {
              setIssuing(false);
            }
          }}
          className="btn-primary mt-4 w-full"
        >
          {issuing && <Loader2 className="h-4 w-4 animate-spin" />}
          Terbitkan ulang
        </button>
      </section>

      {loading && <Spinner />}
      {error && <ErrorNote message={error} onRetry={reload} />}
      {data && data.length === 0 && <EmptyState message="Prize pool masih kosong." />}

      <ul className="space-y-2">
        {data?.map((prize) => (
          <li key={prize.id} className="card-row">
            <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-black text-ink">{prize.nama_hadiah}</p>
                <p className="text-xs text-muted">
                  {prize.alokasi === "undian" ? "Diundi" : "Pemenang kategori"}
                  {prize.donatur_nama ? ` · donasi ${prize.donatur_nama}` : " · seed panitia"}
                </p>
                {prize.pemenang_nama && (
                  <p className="mt-1.5 inline-flex items-center gap-1.5 rounded-lg bg-emerald-50 px-2 py-1 text-xs font-bold text-emerald-700">
                    <Trophy className="h-3.5 w-3.5" />
                    {prize.pemenang_nama}
                  </p>
                )}
              </div>

              <div className="flex flex-wrap items-center gap-2 lg:shrink-0">
                {prize.alokasi === "undian" && !prize.pemenang_id && prize.status === "approved" && (
                  <button
                    type="button"
                    disabled={busy === prize.id}
                    onClick={() => draw(prize.id, prize.nama_hadiah)}
                    className="btn-primary min-h-11 shrink-0 px-3 text-xs"
                  >
                    <Dices className="h-4 w-4" />
                    Undi
                  </button>
                )}
                <select
                  disabled={busy === prize.id}
                  value={prize.status}
                  onChange={(e) => setStatus(prize.id, e.target.value)}
                  className="field min-h-11 w-auto py-0"
                >
                  <option value="pending">Pending</option>
                  <option value="approved">Approved</option>
                  <option value="rejected">Rejected</option>
                  <option value="awarded">Awarded</option>
                </select>
                {busy === prize.id && <Loader2 className="h-4 w-4 animate-spin text-muted" />}
              </div>
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
}

/**
 * Contact spheres are what POWER_TEAM week rewards: a set of classifications
 * that naturally refer each other business. A transaction whose two sides sit
 * in the same sphere scores 1.5×.
 */
function SphereManager() {
  const { data, error, loading, reload } = useApi(() => api.admin.spheres.list());
  const { data: meta } = useApi(() => api.admin.teams.meta());
  const [nama, setNama] = useState("");
  const [selected, setSelected] = useState<string[]>([]);
  const [busy, setBusy] = useState(false);

  const classifications = meta?.classifications ?? [];

  function toggle(id: string) {
    setSelected((prev) => (prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]));
  }

  async function create(e: FormEvent) {
    e.preventDefault();
    if (selected.length < 2) {
      toast.error("Pilih minimal dua klasifikasi — satu sphere butuh dua sisi untuk saling merujuk.");
      return;
    }

    setBusy(true);
    try {
      await api.admin.spheres.create(nama, null, selected);
      setNama("");
      setSelected([]);
      reload();
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Gagal menyimpan.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="space-y-4">
      <form onSubmit={create} className="card space-y-3 p-4">
        <h2 className="flex items-center gap-2 text-base font-black text-ink">
          <Network className="h-[1.1rem] w-[1.1rem] text-brand-600" />
          Buat contact sphere
        </h2>
        <p className="text-xs leading-relaxed text-muted">
          Kelompokkan klasifikasi yang secara alami saling merujuk bisnis — misalnya sphere
          pernikahan berisi Fotografi, Katering, dan Venue.
        </p>

        <input
          required
          value={nama}
          onChange={(e) => setNama(e.target.value)}
          placeholder="Nama sphere, mis. Wedding"
          className="field"
        />

        <fieldset>
          <legend className="section-label mb-2">Klasifikasi anggota</legend>
          <div className="flex flex-wrap gap-2">
            {classifications.map((k) => {
              const on = selected.includes(k.id);
              return (
                <button
                  key={k.id}
                  type="button"
                  onClick={() => toggle(k.id)}
                  aria-pressed={on}
                  className={`min-h-10 rounded-full border px-3 text-xs font-bold transition ${
                    on
                      ? "border-brand-600 bg-brand-600 text-white"
                      : "border-brand-100 bg-white text-muted hover:bg-brand-50"
                  }`}
                >
                  {k.nama}
                </button>
              );
            })}
          </div>
        </fieldset>

        <button type="submit" disabled={busy} className="btn-primary w-full">
          {busy && <Loader2 className="h-4 w-4 animate-spin" />}
          Simpan sphere
        </button>
      </form>

      {loading && <Spinner />}
      {error && <ErrorNote message={error} onRetry={reload} />}
      {data && data.length === 0 && (
        <EmptyState message="Belum ada contact sphere. POWER_TEAM tidak akan memberi bonus sampai ada." />
      )}

      <ul className="space-y-2">
        {data?.map((sphere) => (
          <li key={sphere.id} className="card-row">
            <div className="flex items-start justify-between gap-3">
              <div className="min-w-0 flex-1">
                <p className="text-sm font-black text-ink">{sphere.nama}</p>
                <div className="mt-1.5 flex flex-wrap gap-1.5">
                  {sphere.klasifikasi.map((k) => (
                    <span
                      key={k.id}
                      className="rounded-full bg-brand-50 px-2.5 py-0.5 text-[0.68rem] font-bold text-brand-700"
                    >
                      {k.nama}
                    </span>
                  ))}
                </div>
              </div>
              <button
                type="button"
                aria-label={`Hapus ${sphere.nama}`}
                onClick={async () => {
                  if (!confirm(`Hapus sphere "${sphere.nama}"?`)) return;
                  await api.admin.spheres.remove(sphere.id);
                  reload();
                }}
                className="grid h-10 w-10 shrink-0 place-items-center rounded-lg border border-red-100 text-red-500 transition hover:bg-red-50"
              >
                <Trash2 className="h-3.5 w-3.5" />
              </button>
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
}
