import { useState, type FormEvent } from "react";
import { Ban, KeyRound, Loader2, Users } from "lucide-react";
import { api, ApiError } from "../lib/api";
import { useApi } from "../lib/use-api";
import { formatCurrency, formatDate, today } from "../lib/format";
import { Badge, EmptyState, ErrorNote, PageHeader, Spinner, Tabs } from "../components/ui";

type Tab = "roster" | "tyfcb" | "visitor";

export function CaptainPage() {
  const [tab, setTab] = useState<Tab>("roster");
  const { data, error, loading, reload } = useApi(() => api.captain.team());

  if (loading) return <Spinner />;
  if (error) return <ErrorNote message={error} onRetry={reload} />;
  if (!data) return null;

  return (
    <div className="space-y-5">
      <PageHeader
        eyebrow="Captain Area"
        title="Panel Kapten"
        description="Catat kontribusi atas nama anggota timmu, dan bantu reset kata sandi mereka."
      />

      <Tabs
        tabs={[
          { key: "roster" as Tab, label: `Anggota (${data.members.length})` },
          { key: "tyfcb" as Tab, label: `TYFCB (${data.pending_tyfcb.length})` },
          { key: "visitor" as Tab, label: `Visitor (${data.terdaftar_visitors.length})` },
        ]}
        active={tab}
        onChange={setTab}
      />

      {tab === "roster" && <Roster members={data.members} onChanged={reload} />}
      {tab === "tyfcb" && (
        <TyfcbTab members={data.members} pending={data.pending_tyfcb} onChanged={reload} />
      )}
      {tab === "visitor" && (
        <VisitorTab members={data.members} visitors={data.terdaftar_visitors} onChanged={reload} />
      )}
    </div>
  );
}

function Roster({
  members,
  onChanged,
}: {
  members: { id: string; full_name: string; email: string; color_status: string; role: string }[];
  onChanged: () => void;
}) {
  const [target, setTarget] = useState<string | null>(null);
  const [password, setPassword] = useState("");
  const [busy, setBusy] = useState(false);

  async function reset(id: string) {
    if (password.length < 6) {
      alert("Kata sandi minimal 6 karakter.");
      return;
    }
    setBusy(true);
    try {
      await api.captain.setPassword(id, password);
      setTarget(null);
      setPassword("");
      onChanged();
      alert("Kata sandi berhasil diganti.");
    } catch (err) {
      alert(err instanceof ApiError ? err.message : "Gagal mengganti kata sandi.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <ul className="space-y-2">
      {members.map((member) => (
        <li key={member.id} className="card p-3.5">
          <div className="flex items-center justify-between gap-3">
            <div className="min-w-0">
              <p className="truncate text-sm font-bold text-ink">{member.full_name}</p>
              <p className="truncate text-xs text-muted">{member.email}</p>
            </div>
            <div className="flex shrink-0 items-center gap-2">
              <Badge value={member.color_status} />
              <button
                type="button"
                onClick={() => setTarget(target === member.id ? null : member.id)}
                aria-label={`Reset kata sandi ${member.full_name}`}
                className="grid h-10 w-10 place-items-center rounded-lg border border-brand-100 text-muted transition hover:text-brand-600"
              >
                <KeyRound className="h-3.5 w-3.5" />
              </button>
            </div>
          </div>

          {target === member.id && (
            <div className="mt-3 flex gap-2">
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Kata sandi baru"
                className="field flex-1"
              />
              <button
                type="button"
                disabled={busy}
                onClick={() => void reset(member.id)}
                className="btn-primary shrink-0 px-4"
              >
                {busy && <Loader2 className="h-4 w-4 animate-spin" />}
                Simpan
              </button>
            </div>
          )}
        </li>
      ))}

      {members.length === 0 && <EmptyState message="Tim ini belum punya anggota." />}
    </ul>
  );
}

function TyfcbTab({
  members,
  pending,
  onChanged,
}: {
  members: { id: string; full_name: string }[];
  pending: {
    id: string;
    nilai: number;
    tanggal: string;
    giver_name?: string;
    receiver_name?: string;
  }[];
  onChanged: () => void;
}) {
  const [form, setForm] = useState({ member_id: "", buyer_id: "", nilai: "", tanggal: today() });
  const [busy, setBusy] = useState(false);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    try {
      await api.captain.submitTyfcb(
        form.member_id,
        form.buyer_id,
        Number(form.nilai),
        form.tanggal,
      );
      setForm({ member_id: "", buyer_id: "", nilai: "", tanggal: today() });
      onChanged();
    } catch (err) {
      alert(err instanceof ApiError ? err.message : "Gagal menyimpan.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="space-y-4">
      <form onSubmit={submit} className="card space-y-3 p-4">
        <h2 className="text-base font-black text-ink">Catat TYFCB atas nama anggota</h2>

        <select
          required
          value={form.member_id}
          onChange={(e) => setForm({ ...form, member_id: e.target.value })}
          className="field"
        >
          <option value="">— Penjual (anggota tim Anda)</option>
          {members.map((m) => (
            <option key={m.id} value={m.id}>
              {m.full_name}
            </option>
          ))}
        </select>

        <select
          required
          value={form.buyer_id}
          onChange={(e) => setForm({ ...form, buyer_id: e.target.value })}
          className="field"
        >
          <option value="">— Pembeli (penerima poin)</option>
          {members
            .filter((m) => m.id !== form.member_id)
            .map((m) => (
              <option key={m.id} value={m.id}>
                {m.full_name}
              </option>
            ))}
        </select>

        <input
          type="number"
          min="1"
          required
          value={form.nilai}
          onChange={(e) => setForm({ ...form, nilai: e.target.value })}
          placeholder="Nilai transaksi (IDR)"
          className="field num"
        />

        <input
          type="date"
          required
          value={form.tanggal}
          onChange={(e) => setForm({ ...form, tanggal: e.target.value })}
          className="field"
        />

        <button type="submit" disabled={busy} className="btn-primary w-full">
          {busy && <Loader2 className="h-4 w-4 animate-spin" />}
          Simpan
        </button>
      </form>

      <section>
        <h2 className="section-label mb-2">Pending di tim ini</h2>
        {pending.length === 0 ? (
          <EmptyState message="Tidak ada TYFCB pending." />
        ) : (
          <ul className="space-y-2">
            {pending.map((entry) => (
              <li key={entry.id} className="card flex items-center justify-between gap-3 p-3.5">
                <div className="min-w-0">
                  <p className="truncate text-sm font-bold text-ink">
                    {entry.giver_name} ← {entry.receiver_name}
                  </p>
                  <p className="num text-xs text-muted">
                    {formatCurrency(entry.nilai)} · {formatDate(entry.tanggal)}
                  </p>
                </div>
                <VoidButton onVoid={() => api.captain.voidTyfcb(entry.id)} onDone={onChanged} />
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  );
}

function VisitorTab({
  members,
  visitors,
  onChanged,
}: {
  members: { id: string; full_name: string }[];
  visitors: { id: string; nama: string; kontak: string; inviter_name?: string }[];
  onChanged: () => void;
}) {
  const [form, setForm] = useState({ member_id: "", nama: "", kontak: "", tanggal: today() });
  const [busy, setBusy] = useState(false);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    try {
      await api.captain.registerVisitor(form.member_id, form.nama, form.kontak, form.tanggal);
      setForm({ member_id: "", nama: "", kontak: "", tanggal: today() });
      onChanged();
    } catch (err) {
      alert(err instanceof ApiError ? err.message : "Gagal menyimpan.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="space-y-4">
      <form onSubmit={submit} className="card space-y-3 p-4">
        <h2 className="flex items-center gap-2 text-base font-black text-ink">
          <Users className="h-[1.1rem] w-[1.1rem] text-brand-600" />
          Daftarkan visitor atas nama anggota
        </h2>

        <select
          required
          value={form.member_id}
          onChange={(e) => setForm({ ...form, member_id: e.target.value })}
          className="field"
        >
          <option value="">— Pengundang</option>
          {members.map((m) => (
            <option key={m.id} value={m.id}>
              {m.full_name}
            </option>
          ))}
        </select>

        <input
          required
          value={form.nama}
          onChange={(e) => setForm({ ...form, nama: e.target.value })}
          placeholder="Nama tamu"
          className="field"
        />
        <input
          required
          value={form.kontak}
          onChange={(e) => setForm({ ...form, kontak: e.target.value })}
          placeholder="Kontak"
          className="field num"
        />
        <input
          type="date"
          required
          value={form.tanggal}
          onChange={(e) => setForm({ ...form, tanggal: e.target.value })}
          className="field"
        />

        <button type="submit" disabled={busy} className="btn-primary w-full">
          {busy && <Loader2 className="h-4 w-4 animate-spin" />}
          Simpan
        </button>
      </form>

      <section>
        <h2 className="section-label mb-2">Terdaftar di tim ini</h2>
        {visitors.length === 0 ? (
          <EmptyState message="Belum ada visitor terdaftar." />
        ) : (
          <ul className="space-y-2">
            {visitors.map((visitor) => (
              <li key={visitor.id} className="card flex items-center justify-between gap-3 p-3.5">
                <div className="min-w-0">
                  <p className="truncate text-sm font-bold text-ink">{visitor.nama}</p>
                  <p className="num text-xs text-muted">
                    {visitor.kontak} · {visitor.inviter_name ?? "—"}
                  </p>
                </div>
                <VoidButton onVoid={() => api.captain.voidVisitor(visitor.id)} onDone={onChanged} />
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  );
}

function VoidButton({ onVoid, onDone }: { onVoid: () => Promise<unknown>; onDone: () => void }) {
  const [busy, setBusy] = useState(false);

  return (
    <button
      type="button"
      disabled={busy}
      onClick={async () => {
        if (!confirm("Batalkan entri ini? Poin yang sudah diberikan akan dikembalikan.")) return;
        setBusy(true);
        try {
          await onVoid();
          onDone();
        } catch (err) {
          alert(err instanceof ApiError ? err.message : "Gagal membatalkan.");
        } finally {
          setBusy(false);
        }
      }}
      className="flex min-h-10 shrink-0 items-center gap-1.5 rounded-lg border border-red-200 bg-red-50 px-3 text-xs font-bold text-red-600 transition hover:bg-red-100 disabled:opacity-50"
    >
      {busy ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Ban className="h-3.5 w-3.5" />}
      Void
    </button>
  );
}
