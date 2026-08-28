import { useState, type FormEvent } from "react";
import { CheckCircle2, Clock, Loader2, Pencil, Plus, Trash2, X, XCircle } from "lucide-react";
import { api, ApiError } from "../lib/api";
import { toast } from "../lib/toast-store";
import { useApi } from "../lib/use-api";
import { formatCurrency, formatDate } from "../lib/format";
import { Badge, EmptyState, ErrorNote, PageHeader, Spinner, Tabs } from "../components/ui";
import type { Member, Visitor } from "../lib/types";

// ── TYFCB verification ────────────────────────────────────────────────────

type TyfcbTab = "pending" | "" | "verified" | "rejected";

export function AdminTyfcbPage() {
  const [tab, setTab] = useState<TyfcbTab>("pending");
  const [busy, setBusy] = useState<string | null>(null);
  const { data, error, loading, reload } = useApi(() => api.admin.tyfcb.list(tab), [tab]);

  async function setStatus(id: string, status: string) {
    setBusy(id + status);
    try {
      await api.admin.tyfcb.setStatus(id, status);
      reload();
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Gagal memperbarui status.");
    } finally {
      setBusy(null);
    }
  }

  const counts = data?.counts ?? {};

  return (
    <div className="space-y-5">
      <PageHeader
        eyebrow="Admin Area"
        title="Verifikasi TYFCB"
        description="Approve atau reject submission TYFCB dari member. Poin hanya masuk ledger saat berstatus verified."
      />

      <Tabs
        tabs={[
          { key: "pending" as TyfcbTab, label: `Pending (${counts.pending ?? 0})` },
          { key: "" as TyfcbTab, label: "Semua" },
          { key: "verified" as TyfcbTab, label: `Verified (${counts.verified ?? 0})` },
          { key: "rejected" as TyfcbTab, label: `Rejected (${counts.rejected ?? 0})` },
        ]}
        active={tab}
        onChange={setTab}
      />

      {loading && <Spinner />}
      {error && <ErrorNote message={error} onRetry={reload} />}
      {data && data.entries.length === 0 && <EmptyState message="Tidak ada submission." />}

      <div className="space-y-2.5">
        {data?.entries.map((entry) => (
          <div key={entry.id} className="card p-3.5">
            <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
              <div className="min-w-0 flex-1">
                <div className="flex flex-wrap items-center gap-x-2 gap-y-1">
                  <Badge value={entry.status} />
                  <p className="num text-sm font-black text-ink">{formatCurrency(entry.nilai)}</p>
                  <span className="text-xs text-muted">· {formatDate(entry.tanggal)}</span>
                  {entry.computed_score != null && (
                    <span className="num text-xs font-bold text-brand-700">
                      +{entry.computed_score} pts
                    </span>
                  )}
                </div>

                <dl className="mt-2 space-y-0.5 text-sm">
                  <div className="flex gap-2">
                    <dt className="w-14 shrink-0 text-xs leading-5 text-muted">Pembeli</dt>
                    <dd className="min-w-0 truncate font-bold text-ink">{entry.giver_name ?? "—"}</dd>
                  </div>
                  <div className="flex gap-2">
                    <dt className="w-14 shrink-0 text-xs leading-5 text-muted">Penjual</dt>
                    <dd className="min-w-0 truncate font-bold text-ink">
                      {entry.receiver_name ?? "—"}
                    </dd>
                  </div>
                </dl>
              </div>

              <div className="grid grid-cols-2 gap-2 lg:flex lg:shrink-0">
                {entry.status !== "verified" && (
                  <ActionButton
                    busy={busy === entry.id + "verified"}
                    onClick={() => setStatus(entry.id, "verified")}
                    icon={CheckCircle2}
                    label="Approve"
                    className="bg-emerald-600 text-white hover:bg-emerald-700"
                  />
                )}
                {entry.status !== "pending" && (
                  <ActionButton
                    busy={busy === entry.id + "pending"}
                    onClick={() => setStatus(entry.id, "pending")}
                    icon={Clock}
                    label="Pending"
                    className="border border-amber-300 bg-amber-50 text-amber-700 hover:bg-amber-100"
                  />
                )}
                {entry.status !== "rejected" && (
                  <ActionButton
                    busy={busy === entry.id + "rejected"}
                    onClick={() => setStatus(entry.id, "rejected")}
                    icon={XCircle}
                    label="Reject"
                    className="border border-red-200 bg-red-50 text-red-600 hover:bg-red-100"
                  />
                )}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function ActionButton({
  busy,
  onClick,
  icon: Icon,
  label,
  className,
}: {
  busy: boolean;
  onClick: () => void;
  icon: typeof CheckCircle2;
  label: string;
  className: string;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={busy}
      className={`flex min-h-11 items-center justify-center gap-1.5 rounded-xl px-3 text-xs font-black transition disabled:opacity-50 ${className}`}
    >
      {busy ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Icon className="h-3.5 w-3.5" />}
      {label}
    </button>
  );
}

// ── Visitors ──────────────────────────────────────────────────────────────

export function AdminVisitorsPage() {
  const [tab, setTab] = useState("");
  const [busy, setBusy] = useState<string | null>(null);
  const { data, error, loading, reload } = useApi(() => api.admin.visitors.list(tab), [tab]);

  async function update(id: string, body: { status_hadir?: string; is_converted?: boolean }) {
    setBusy(id);
    try {
      await api.admin.visitors.update(id, body);
      reload();
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Gagal memperbarui visitor.");
    } finally {
      setBusy(null);
    }
  }

  return (
    <div className="space-y-5">
      <PageHeader
        eyebrow="Admin Area"
        title="Kelola Visitor"
        description="Poin visitor bertingkat: hadir 20, hadir penuh 50, konversi +100. Menurunkan status membalik poinnya."
      />

      <Tabs
        tabs={[
          { key: "", label: "Semua" },
          { key: "terdaftar", label: "Terdaftar" },
          { key: "hadir", label: "Hadir" },
          { key: "hadir_penuh", label: "Hadir Penuh" },
        ]}
        active={tab}
        onChange={setTab}
      />

      {loading && <Spinner />}
      {error && <ErrorNote message={error} onRetry={reload} />}
      {data && data.length === 0 && <EmptyState message="Belum ada visitor." />}

      <div className="space-y-2.5">
        {data?.map((visitor) => (
          <VisitorRow
            key={visitor.id}
            visitor={visitor}
            busy={busy === visitor.id}
            onUpdate={(body) => update(visitor.id, body)}
          />
        ))}
      </div>
    </div>
  );
}

function VisitorRow({
  visitor,
  busy,
  onUpdate,
}: {
  visitor: Visitor;
  busy: boolean;
  onUpdate: (body: { status_hadir?: string; is_converted?: boolean }) => void;
}) {
  return (
    <div className="card p-3.5">
      <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <p className="truncate text-sm font-black text-ink">{visitor.nama}</p>
            <Badge value={visitor.status_hadir} />
            {visitor.is_converted && <Badge value="verified" />}
          </div>
          <p className="num mt-1 text-xs text-muted">
            {visitor.kontak} · diundang {visitor.inviter_name ?? "—"} ·{" "}
            {formatDate(visitor.tanggal_undang)}
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-2 lg:shrink-0">
          <select
            disabled={busy || visitor.is_void}
            value={visitor.status_hadir}
            onChange={(e) => onUpdate({ status_hadir: e.target.value })}
            className="field min-h-11 w-auto py-0"
          >
            <option value="terdaftar">Terdaftar</option>
            <option value="hadir">Hadir</option>
            <option value="hadir_penuh">Hadir Penuh</option>
          </select>

          <label className="flex min-h-11 items-center gap-2 rounded-xl border border-brand-100 px-3 text-xs font-bold text-ink">
            <input
              type="checkbox"
              disabled={busy || visitor.is_void}
              checked={visitor.is_converted}
              onChange={(e) => onUpdate({ is_converted: e.target.checked })}
              className="h-4 w-4 accent-brand-600"
            />
            Konversi
          </label>

          {busy && <Loader2 className="h-4 w-4 animate-spin text-muted" />}
        </div>
      </div>
    </div>
  );
}

// ── Members ───────────────────────────────────────────────────────────────

export function AdminMembersPage() {
  const [editing, setEditing] = useState<Member | null>(null);
  const [creating, setCreating] = useState(false);

  const { data, error, loading, reload } = useApi(() => api.admin.members.list());
  const { data: meta } = useApi(() => api.admin.teams.meta());

  const grouped = groupByTeam(data ?? []);

  return (
    <div className="space-y-5">
      <PageHeader
        eyebrow="Admin Area"
        title="Kelola Member"
        description={`${data?.length ?? 0} member terdaftar di season ini.`}
        action={
          <button type="button" onClick={() => setCreating(true)} className="btn-primary shrink-0">
            <Plus className="h-4 w-4" />
            Tambah
          </button>
        }
      />

      {loading && <Spinner />}
      {error && <ErrorNote message={error} onRetry={reload} />}

      {grouped.map(([teamName, members]) => (
        <section key={teamName} className="card overflow-hidden">
          <div className="flex items-center justify-between border-b border-brand-50 px-4 py-3">
            <h2 className="font-black text-ink">{teamName}</h2>
            <span className="text-sm text-muted">{members.length} member</span>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-brand-50 text-xs font-bold uppercase tracking-wide text-muted">
                  <th className="px-2 py-2 text-left sm:px-3">Nama / Email</th>
                  <th className="hidden px-3 py-2 text-left sm:table-cell">Klasifikasi</th>
                  <th className="px-2 py-2 text-left sm:px-3">Status</th>
                  <th className="px-2 py-2 text-left sm:px-3">Role</th>
                  <th className="px-2 py-2 sm:px-3" />
                </tr>
              </thead>
              <tbody className="divide-y divide-brand-50">
                {members.map((member) => (
                  <tr key={member.id}>
                    <td className="px-2 py-2.5 sm:px-3">
                      <p className="font-bold text-ink">{member.full_name}</p>
                      <p className="text-xs text-muted">{member.email}</p>
                    </td>
                    <td className="hidden px-3 py-2.5 text-muted sm:table-cell">
                      {member.klasifikasi_nama ?? "—"}
                    </td>
                    <td className="px-2 py-2.5 sm:px-3">
                      <Badge value={member.color_status} />
                    </td>
                    <td className="px-2 py-2.5 sm:px-3">
                      <Badge value={member.role} />
                    </td>
                    <td className="px-2 py-2.5 text-right sm:px-3">
                      <button
                        type="button"
                        onClick={() => setEditing(member)}
                        aria-label={`Edit ${member.full_name}`}
                        className="inline-flex min-h-9 items-center gap-1 rounded-lg border border-brand-100 px-2.5 text-xs font-bold text-muted transition hover:text-brand-600"
                      >
                        <Pencil className="h-3 w-3" />
                        <span className="hidden sm:inline">Edit</span>
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>
      ))}

      {(editing || creating) && (
        <MemberDialog
          member={editing}
          teams={meta?.teams ?? []}
          classifications={meta?.classifications ?? []}
          onClose={() => {
            setEditing(null);
            setCreating(false);
          }}
          onSaved={() => {
            setEditing(null);
            setCreating(false);
            reload();
          }}
        />
      )}
    </div>
  );
}

function groupByTeam(members: Member[]): [string, Member[]][] {
  const map = new Map<string, Member[]>();
  for (const member of members) {
    const key = member.nama_tim ?? "Tanpa tim";
    map.set(key, [...(map.get(key) ?? []), member]);
  }
  return [...map.entries()];
}

function MemberDialog({
  member,
  teams,
  classifications,
  onClose,
  onSaved,
}: {
  member: Member | null;
  teams: { id: string; nama_tim: string }[];
  classifications: { id: string; nama: string }[];
  onClose: () => void;
  onSaved: () => void;
}) {
  const [form, setForm] = useState({
    full_name: member?.full_name ?? "",
    email: member?.email ?? "",
    password: "",
    team_id: member?.team_id ?? "",
    klasifikasi_id: member?.klasifikasi_id ?? "",
    color_status: member?.color_status ?? "merah",
    role: member?.role ?? "member",
    is_active: member?.is_active ?? true,
  });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setSaving(true);

    try {
      if (member) {
        await api.admin.members.update(member.id, {
          full_name: form.full_name,
          email: form.email,
          team_id: form.team_id,
          klasifikasi_id: form.klasifikasi_id,
          color_status: form.color_status,
          role: form.role,
          is_active: form.is_active,
          ...(form.password ? { new_password: form.password } : {}),
        });
      } else {
        await api.admin.members.create({
          full_name: form.full_name,
          email: form.email,
          password: form.password,
          team_id: form.team_id || null,
          klasifikasi_id: form.klasifikasi_id || null,
          color_status: form.color_status,
          role: form.role,
        });
      }
      onSaved();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Gagal menyimpan.");
    } finally {
      setSaving(false);
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-end justify-center bg-ink/40 px-3 pb-3 backdrop-blur-sm sm:items-center sm:p-4"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <form
        onSubmit={submit}
        className="max-h-[85vh] w-full max-w-md overflow-y-auto rounded-3xl bg-white p-5 shadow-2xl"
      >
        <div className="mb-4 flex items-center justify-between gap-3">
          <h3 className="text-lg font-black text-ink">
            {member ? "Edit Member" : "Tambah Member"}
          </h3>
          <button
            type="button"
            onClick={onClose}
            aria-label="Tutup"
            className="grid h-10 w-10 shrink-0 place-items-center rounded-full text-muted transition hover:bg-brand-50"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="space-y-3">
          <Field label="Nama lengkap">
            <input
              required
              value={form.full_name}
              onChange={(e) => setForm({ ...form, full_name: e.target.value })}
              className="field"
            />
          </Field>

          <Field label="Email">
            <input
              type="email"
              required
              value={form.email}
              onChange={(e) => setForm({ ...form, email: e.target.value })}
              className="field"
            />
          </Field>

          <Field label={member ? "Kata sandi baru (opsional)" : "Kata sandi"}>
            <input
              type="password"
              required={!member}
              minLength={member ? 0 : 6}
              value={form.password}
              onChange={(e) => setForm({ ...form, password: e.target.value })}
              placeholder={member ? "Kosongkan jika tidak diubah" : ""}
              className="field"
            />
          </Field>

          <Field label="Tim">
            <select
              value={form.team_id}
              onChange={(e) => setForm({ ...form, team_id: e.target.value })}
              className="field"
            >
              <option value="">— Tanpa tim</option>
              {teams.map((team) => (
                <option key={team.id} value={team.id}>
                  {team.nama_tim}
                </option>
              ))}
            </select>
          </Field>

          <Field label="Klasifikasi">
            <select
              value={form.klasifikasi_id}
              onChange={(e) => setForm({ ...form, klasifikasi_id: e.target.value })}
              className="field"
            >
              <option value="">— Pilih</option>
              {classifications.map((item) => (
                <option key={item.id} value={item.id}>
                  {item.nama}
                </option>
              ))}
            </select>
          </Field>

          <div className="grid grid-cols-2 gap-3">
            <Field label="Status warna">
              <select
                value={form.color_status}
                onChange={(e) =>
                  setForm({ ...form, color_status: e.target.value as typeof form.color_status })
                }
                className="field capitalize"
              >
                <option value="merah">Merah</option>
                <option value="kuning">Kuning</option>
                <option value="hijau">Hijau</option>
              </select>
            </Field>

            <Field label="Role">
              <select
                value={form.role}
                onChange={(e) => setForm({ ...form, role: e.target.value as typeof form.role })}
                className="field"
              >
                <option value="member">Member</option>
                <option value="captain">Kapten</option>
                <option value="admin">Admin</option>
              </select>
            </Field>
          </div>

          {member && (
            <label className="flex min-h-11 items-center gap-2 text-sm font-bold text-ink">
              <input
                type="checkbox"
                checked={form.is_active}
                onChange={(e) => setForm({ ...form, is_active: e.target.checked })}
                className="h-4 w-4 accent-brand-600"
              />
              Aktif
            </label>
          )}
        </div>

        {error && (
          <p role="alert" className="mt-3 rounded-xl bg-red-50 px-3 py-2 text-sm font-semibold text-red-700">
            {error}
          </p>
        )}

        <div className="mt-5 flex gap-2">
          <button type="button" onClick={onClose} className="btn-secondary flex-1">
            Batal
          </button>
          <button type="submit" disabled={saving} className="btn-primary flex-1">
            {saving && <Loader2 className="h-4 w-4 animate-spin" />}
            Simpan
          </button>
        </div>
      </form>
    </div>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="block">
      <span className="section-label mb-1.5 block">{label}</span>
      {children}
    </label>
  );
}

// ── Simple master-data screens (teams, classifications) ───────────────────

export function AdminTeamsPage() {
  return (
    <MasterData
      title="Kelola Tim"
      description="Tim yang berlaga di season ini."
      load={() => api.admin.teams.list().then((rows) => rows.map((t) => ({ id: t.id, nama: t.nama_tim, meta: `${t.member_count ?? 0} member` })))}
      create={(nama) => api.admin.teams.create(nama)}
      rename={(id, nama) => api.admin.teams.rename(id, nama)}
      remove={(id) => api.admin.teams.remove(id)}
    />
  );
}

export function AdminClassificationsPage() {
  return (
    <MasterData
      title="Klasifikasi Bisnis"
      description="Kategori bisnis yang bisa dipilih member."
      load={() => api.admin.classifications.list().then((rows) => rows.map((c) => ({ id: c.id, nama: c.nama })))}
      create={(nama) => api.admin.classifications.create(nama)}
      rename={(id, nama) => api.admin.classifications.rename(id, nama)}
      remove={(id) => api.admin.classifications.remove(id)}
    />
  );
}

type MasterRow = { id: string; nama: string; meta?: string };

function MasterData({
  title,
  description,
  load,
  create,
  rename,
  remove,
}: {
  title: string;
  description: string;
  load: () => Promise<MasterRow[]>;
  create: (nama: string) => Promise<unknown>;
  rename: (id: string, nama: string) => Promise<unknown>;
  remove: (id: string) => Promise<unknown>;
}) {
  const { data, error, loading, reload } = useApi(load);
  const [draft, setDraft] = useState("");
  const [editing, setEditing] = useState<{ id: string; nama: string } | null>(null);
  const [busy, setBusy] = useState(false);

  async function run(action: () => Promise<unknown>) {
    setBusy(true);
    try {
      await action();
      reload();
      setDraft("");
      setEditing(null);
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Gagal menyimpan.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="space-y-5">
      <PageHeader eyebrow="Admin Area" title={title} description={description} />

      <form
        onSubmit={(e) => {
          e.preventDefault();
          if (draft.trim()) void run(() => create(draft.trim()));
        }}
        className="card flex gap-2 p-3"
      >
        <input
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          placeholder="Nama baru"
          className="field flex-1"
        />
        <button type="submit" disabled={busy || !draft.trim()} className="btn-primary shrink-0">
          <Plus className="h-4 w-4" />
          Tambah
        </button>
      </form>

      {loading && <Spinner />}
      {error && <ErrorNote message={error} onRetry={reload} />}
      {data && data.length === 0 && <EmptyState message="Belum ada data." />}

      <ul className="space-y-2">
        {data?.map((row) => (
          <li key={row.id} className="card flex items-center gap-3 p-3">
            {editing?.id === row.id ? (
              <>
                <input
                  value={editing.nama}
                  onChange={(e) => setEditing({ ...editing, nama: e.target.value })}
                  className="field flex-1"
                />
                <button
                  type="button"
                  disabled={busy}
                  onClick={() => void run(() => rename(row.id, editing.nama))}
                  className="btn-primary shrink-0 px-4"
                >
                  Simpan
                </button>
                <button
                  type="button"
                  onClick={() => setEditing(null)}
                  className="btn-secondary shrink-0 px-4"
                >
                  Batal
                </button>
              </>
            ) : (
              <>
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-bold text-ink">{row.nama}</p>
                  {row.meta && <p className="text-xs text-muted">{row.meta}</p>}
                </div>
                <button
                  type="button"
                  onClick={() => setEditing({ id: row.id, nama: row.nama })}
                  aria-label={`Ubah ${row.nama}`}
                  className="grid h-10 w-10 shrink-0 place-items-center rounded-lg border border-brand-100 text-muted transition hover:text-brand-600"
                >
                  <Pencil className="h-3.5 w-3.5" />
                </button>
                <button
                  type="button"
                  disabled={busy}
                  onClick={() => {
                    if (confirm(`Hapus "${row.nama}"?`)) void run(() => remove(row.id));
                  }}
                  aria-label={`Hapus ${row.nama}`}
                  className="grid h-10 w-10 shrink-0 place-items-center rounded-lg border border-red-100 text-red-500 transition hover:bg-red-50"
                >
                  <Trash2 className="h-3.5 w-3.5" />
                </button>
              </>
            )}
          </li>
        ))}
      </ul>
    </div>
  );
}
