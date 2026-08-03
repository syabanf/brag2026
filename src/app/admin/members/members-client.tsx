"use client";

import { Check, ChevronDown, ChevronUp, Eye, EyeOff, Pencil, Shield, ShieldCheck, ShieldOff, X } from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect, useRef, useState } from "react";

type Member = {
  id: string;
  user_id: string;
  full_name: string;
  email: string;
  role: string;
  team_id: string | null;
  nama_tim: string | null;
  klasifikasi_id: string | null;
  klasifikasi_nama: string | null;
  color_status: string;
  is_active: boolean;
};

type Team = { id: string; nama_tim: string };
type Klas = { id: string; nama: string };

const COLOR_STYLE: Record<string, string> = {
  merah: "bg-red-50 text-red-700",
  kuning: "bg-yellow-50 text-yellow-700",
  hijau: "bg-green-50 text-green-700",
};

function EditRow({
  member,
  teams,
  klasifikasi,
  onCancel,
  onSaved,
}: {
  member: Member;
  teams: Team[];
  klasifikasi: Klas[];
  onCancel: () => void;
  onSaved: () => void;
}) {
  const [form, setForm] = useState({
    full_name: member.full_name,
    email: member.email,
    team_id: member.team_id ?? "",
    klasifikasi_id: member.klasifikasi_id ?? "",
    color_status: member.color_status,
    is_active: member.is_active,
    new_password: "",
  });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showPassword, setShowPassword] = useState(false);

  async function save() {
    if (form.new_password && form.new_password.length < 6) {
      setError("Password baru minimal 6 karakter.");
      return;
    }
    setSaving(true);
    setError(null);
    const res = await fetch(`/api/admin/members/${member.id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(form),
    });
    setSaving(false);
    if (!res.ok) {
      const data = await res.json().catch(() => null);
      setError(data?.error ?? "Gagal menyimpan perubahan.");
      return;
    }
    onSaved();
  }

  return (
    <tr className="bg-brand-50">
      <td className="px-3 py-2">
        <input
          className="w-full rounded-lg border border-brand-200 bg-white px-2 py-1 text-sm font-bold"
          value={form.full_name}
          onChange={(e) => setForm({ ...form, full_name: e.target.value })}
        />
        <input
          className="mt-1 w-full rounded-lg border border-brand-100 bg-white px-2 py-1 text-xs text-muted"
          value={form.email}
          onChange={(e) => setForm({ ...form, email: e.target.value })}
        />
        <div className="relative mt-1">
          <input
            type={showPassword ? "text" : "password"}
            autoComplete="new-password"
            placeholder="Password baru (kosongkan jika tetap)"
            className="w-full rounded-lg border border-brand-100 bg-white px-2 py-1 pr-8 text-xs"
            value={form.new_password}
            onChange={(e) => setForm({ ...form, new_password: e.target.value })}
          />
          <button
            type="button"
            onClick={() => setShowPassword((v) => !v)}
            aria-label={showPassword ? "Sembunyikan password" : "Lihat password"}
            className="absolute right-2 top-1/2 -translate-y-1/2 text-muted hover:text-ink"
          >
            {showPassword ? <EyeOff className="h-3.5 w-3.5" /> : <Eye className="h-3.5 w-3.5" />}
          </button>
        </div>
        {error && <p className="mt-1 text-xs font-bold text-red-600">{error}</p>}
      </td>
      <td className="px-3 py-2">
        <div className="relative">
          <select
            className="w-full appearance-none rounded-lg border border-brand-200 bg-white px-2 py-1 text-sm"
            value={form.team_id}
            onChange={(e) => setForm({ ...form, team_id: e.target.value })}
          >
            <option value="">— Tanpa tim</option>
            {teams.map((t) => (
              <option key={t.id} value={t.id}>{t.nama_tim}</option>
            ))}
          </select>
          <ChevronDown className="pointer-events-none absolute right-2 top-1.5 h-3.5 w-3.5 text-muted" />
        </div>
      </td>
      <td className="px-3 py-2">
        <div className="relative">
          <select
            className="w-full appearance-none rounded-lg border border-brand-200 bg-white px-2 py-1 text-sm"
            value={form.klasifikasi_id}
            onChange={(e) => setForm({ ...form, klasifikasi_id: e.target.value })}
          >
            <option value="">— Pilih</option>
            {klasifikasi.map((k) => (
              <option key={k.id} value={k.id}>{k.nama}</option>
            ))}
          </select>
          <ChevronDown className="pointer-events-none absolute right-2 top-1.5 h-3.5 w-3.5 text-muted" />
        </div>
      </td>
      <td className="px-3 py-2">
        <div className="relative">
          <select
            className="w-full appearance-none rounded-lg border border-brand-200 bg-white px-2 py-1 text-sm"
            value={form.color_status}
            onChange={(e) => setForm({ ...form, color_status: e.target.value })}
          >
            <option value="merah">Merah</option>
            <option value="kuning">Kuning</option>
            <option value="hijau">Hijau</option>
          </select>
          <ChevronDown className="pointer-events-none absolute right-2 top-1.5 h-3.5 w-3.5 text-muted" />
        </div>
      </td>
      <td className="px-3 py-2 text-center">
        <input
          type="checkbox"
          checked={form.is_active}
          onChange={(e) => setForm({ ...form, is_active: e.target.checked })}
          className="h-4 w-4 accent-brand-600"
        />
      </td>
      <td className="px-3 py-2" colSpan={2}>
        <div className="flex gap-1.5">
          <button
            onClick={save}
            disabled={saving}
            className="flex items-center gap-1 rounded-lg bg-brand-600 px-3 py-1.5 text-xs font-bold text-white hover:bg-brand-700 disabled:opacity-50"
          >
            <Check className="h-3.5 w-3.5" />
            {saving ? "..." : "Simpan"}
          </button>
          <button
            onClick={onCancel}
            className="flex items-center gap-1 rounded-lg border border-brand-100 bg-white px-3 py-1.5 text-xs font-bold text-muted hover:text-ink"
          >
            <X className="h-3.5 w-3.5" />
          </button>
        </div>
      </td>
    </tr>
  );
}

const ROLE_META: Record<string, { label: string; badge: string; icon: React.ReactNode }> = {
  admin:   { label: "Admin",   badge: "bg-brand-50 text-brand-700",   icon: <ShieldCheck className="h-3 w-3" /> },
  captain: { label: "Kapten",  badge: "bg-amber-50 text-amber-700",   icon: <Shield className="h-3 w-3" /> },
  member:  { label: "Member",  badge: "bg-slate-50 text-slate-500",   icon: <ShieldOff className="h-3 w-3" /> },
};

const ALL_ROLES = ["admin", "captain", "member"] as const;
type RoleOption = typeof ALL_ROLES[number];

function RolePickerButton({ member, onDone }: { member: Member; onDone: () => void }) {
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function onClickOutside(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    }
    document.addEventListener("mousedown", onClickOutside);
    return () => document.removeEventListener("mousedown", onClickOutside);
  }, []);

  async function changeRole(newRole: RoleOption) {
    setOpen(false);
    const meta = ROLE_META[newRole];
    if (!window.confirm(`Ubah role ${member.full_name} menjadi ${meta.label}?`)) return;

    setLoading(true);
    const res = await fetch(`/api/admin/members/${member.id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ role: newRole }),
    });
    setLoading(false);

    if (!res.ok) {
      const data = await res.json();
      alert(data.error ?? "Gagal mengubah role.");
      return;
    }
    onDone();
  }

  const current = member.role in ROLE_META ? member.role : "member";
  const meta = ROLE_META[current];

  return (
    <div className="relative" ref={ref}>
      <button
        onClick={() => setOpen((v) => !v)}
        disabled={loading}
        className="flex items-center gap-1 rounded-lg border border-brand-100 bg-white px-2.5 py-1 text-xs font-bold text-muted transition hover:border-brand-300 hover:text-ink disabled:opacity-50"
      >
        {loading ? "..." : <>{open ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />} Role</>}
      </button>

      {open && (
        <div className="absolute right-0 top-7 z-20 w-36 overflow-hidden rounded-xl border border-brand-100 bg-white shadow-lg">
          {ALL_ROLES.filter((r) => r !== current).map((r) => {
            const m = ROLE_META[r];
            return (
              <button
                key={r}
                onClick={() => changeRole(r)}
                className="flex w-full items-center gap-2 px-3 py-2 text-xs font-bold text-ink hover:bg-brand-50"
              >
                {m.icon}
                Jadikan {m.label}
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}

export function MemberTable({
  members,
  teams,
  klasifikasi,
}: {
  members: Member[];
  teams: Team[];
  klasifikasi: Klas[];
}) {
  const router = useRouter();
  const [editingId, setEditingId] = useState<string | null>(null);
  const [filterTeam, setFilterTeam] = useState("");

  const filtered = filterTeam
    ? members.filter((m) => m.team_id === filterTeam)
    : members;

  const groups = teams
    .map((t) => ({
      team: t,
      rows: filtered.filter((m) => m.team_id === t.id),
    }))
    .filter((g) => g.rows.length > 0);

  return (
    <div className="space-y-6">
      {/* Filter bar */}
      <div className="flex flex-wrap gap-3">
        <div className="relative">
          <select
            className="appearance-none rounded-full border border-brand-100 bg-white pl-4 pr-8 py-2 text-sm font-bold text-ink focus:outline-none focus:ring-2 focus:ring-brand-500"
            value={filterTeam}
            onChange={(e) => setFilterTeam(e.target.value)}
          >
            <option value="">Semua Team</option>
            {teams.map((t) => (
              <option key={t.id} value={t.id}>{t.nama_tim}</option>
            ))}
          </select>
          <ChevronDown className="pointer-events-none absolute right-3 top-2.5 h-4 w-4 text-muted" />
        </div>
        <span className="self-center text-sm text-muted">{filtered.length} member ditampilkan</span>
      </div>

      {/* Tables per team */}
      {groups.map(({ team, rows }) => (
        <section key={team.id} className="glass-panel overflow-hidden rounded-2xl">
          <div className="flex items-center justify-between border-b border-brand-100 px-4 py-3">
            <h2 className="font-black text-ink">{team.nama_tim}</h2>
            <span className="text-sm text-muted">{rows.length} member</span>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-brand-50 text-xs font-bold uppercase tracking-wide text-muted">
                  <th className="px-3 py-2 text-left">Nama / Email</th>
                  <th className="px-3 py-2 text-left">Team</th>
                  <th className="px-3 py-2 text-left">Klasifikasi</th>
                  <th className="px-3 py-2 text-left">Status</th>
                  <th className="px-3 py-2 text-center">Aktif</th>
                  <th className="px-3 py-2 text-left">Role</th>
                  <th className="px-3 py-2" />
                </tr>
              </thead>
              <tbody className="divide-y divide-brand-50">
                {rows.map((m) =>
                  editingId === m.id ? (
                    <EditRow
                      key={m.id}
                      member={m}
                      teams={teams}
                      klasifikasi={klasifikasi}
                      onCancel={() => setEditingId(null)}
                      onSaved={() => {
                        setEditingId(null);
                        router.refresh();
                      }}
                    />
                  ) : (
                    <tr key={m.id} className="hover:bg-brand-50/50">
                      <td className="px-3 py-2.5">
                        <p className="font-bold text-ink">{m.full_name}</p>
                        <p className="text-xs text-muted">{m.email}</p>
                      </td>
                      <td className="px-3 py-2.5 text-muted">{m.nama_tim ?? "—"}</td>
                      <td className="px-3 py-2.5 text-muted">{m.klasifikasi_nama ?? "—"}</td>
                      <td className="px-3 py-2.5">
                        <span className={`rounded-full px-2.5 py-0.5 text-xs font-bold capitalize ${COLOR_STYLE[m.color_status]}`}>
                          {m.color_status}
                        </span>
                      </td>
                      <td className="px-3 py-2.5 text-center">
                        <span className={`text-xs font-bold ${m.is_active ? "text-green-600" : "text-muted line-through"}`}>
                          {m.is_active ? "✓" : "✗"}
                        </span>
                      </td>
                      <td className="px-3 py-2.5">
                        {(() => {
                          const roleMeta = ROLE_META[m.role] ?? ROLE_META.member;
                          return (
                            <span className={`rounded-full px-2.5 py-0.5 text-xs font-bold ${roleMeta.badge}`}>
                              {roleMeta.label}
                            </span>
                          );
                        })()}
                      </td>
                      <td className="px-3 py-2.5">
                        <div className="flex items-center gap-1.5">
                          <button
                            onClick={() => setEditingId(m.id)}
                            className="flex items-center gap-1 rounded-lg border border-brand-100 bg-white px-2.5 py-1 text-xs font-bold text-muted transition hover:border-brand-300 hover:text-brand-600"
                          >
                            <Pencil className="h-3 w-3" />
                            Edit
                          </button>
                          <RolePickerButton
                            member={m}
                            onDone={() => router.refresh()}
                          />
                        </div>
                      </td>
                    </tr>
                  )
                )}
              </tbody>
            </table>
          </div>
        </section>
      ))}
    </div>
  );
}
