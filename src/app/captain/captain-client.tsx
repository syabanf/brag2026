"use client";

import { Banknote, Check, Eye, EyeOff, KeyRound, Trash2, UserPlus, Users, X } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { formatCurrency } from "@/lib/utils";
import type { PendingTyfcb, SeasonMember, TeamMember, TerdaftarVisitor } from "./page";

const COLOR_BADGE: Record<string, string> = {
  merah:  "bg-red-50 text-red-700",
  kuning: "bg-yellow-50 text-yellow-700",
  hijau:  "bg-green-50 text-green-700",
};

const TABS = [
  { key: "tyfcb",   label: "Kirim TYFCB",   icon: Banknote },
  { key: "visitor", label: "Kirim Visitor",  icon: UserPlus },
  { key: "members", label: "Anggota Tim",    icon: Users    },
] as const;
type TabKey = typeof TABS[number]["key"];

// ─── TYFCB Form ─────────────────────────────────────────────────────────────

function TyfcbForm({
  members,
  allMembers,
}: {
  members: TeamMember[];
  allMembers: SeasonMember[];
}) {
  const router = useRouter();
  const [form, setForm] = useState({ member_id: "", buyer_id: "", nilai: "", tanggal: "" });
  const [loading, setLoading] = useState(false);
  const [msg, setMsg] = useState<{ ok: boolean; text: string } | null>(null);

  const buyers = allMembers.filter((m) => m.id !== form.member_id);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setMsg(null);
    setLoading(true);

    const res = await fetch("/api/captain/tyfcb", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        member_id: form.member_id,
        buyer_id:  form.buyer_id,
        nilai:     Number(form.nilai),
        tanggal:   form.tanggal,
      }),
    });

    const data = await res.json();
    setLoading(false);

    if (!res.ok) {
      setMsg({ ok: false, text: data.error ?? "Gagal menyimpan." });
      return;
    }

    setMsg({ ok: true, text: `TYFCB berhasil dikirim. Skor estimasi: ${data.computed_score} poin.` });
    setForm({ member_id: "", buyer_id: "", nilai: "", tanggal: "" });
    router.refresh();
  }

  return (
    <form onSubmit={submit} className="space-y-4 max-w-lg">
      {msg && (
        <div className={`rounded-2xl border px-4 py-3 text-sm font-semibold ${
          msg.ok ? "border-green-200 bg-green-50 text-green-700" : "border-brand-100 bg-brand-50 text-brand-700"
        }`}>
          {msg.text}
        </div>
      )}

      <label className="block">
        <span className="mb-1.5 block text-sm font-bold text-ink">Dari (anggota tim)</span>
        <select
          required
          className="w-full rounded-2xl border border-brand-100 bg-white px-4 py-3 text-sm font-semibold focus:outline-none focus:ring-2 focus:ring-brand-500"
          value={form.member_id}
          onChange={(e) => setForm({ ...form, member_id: e.target.value, buyer_id: "" })}
        >
          <option value="">— Pilih anggota</option>
          {members.map((m) => (
            <option key={m.id} value={m.id}>{m.full_name}</option>
          ))}
        </select>
      </label>

      <label className="block">
        <span className="mb-1.5 block text-sm font-bold text-ink">Pemberi Bisnis (Pembeli) — mendapat poin</span>
        <select
          required
          disabled={!form.member_id}
          className="w-full rounded-2xl border border-brand-100 bg-white px-4 py-3 text-sm font-semibold focus:outline-none focus:ring-2 focus:ring-brand-500 disabled:opacity-50"
          value={form.buyer_id}
          onChange={(e) => setForm({ ...form, buyer_id: e.target.value })}
        >
          <option value="">— Pilih pembeli</option>
          {buyers.map((m) => (
            <option key={m.id} value={m.id}>
              {m.full_name}{m.nama_tim ? ` · ${m.nama_tim}` : ""}
            </option>
          ))}
        </select>
      </label>

      <label className="block">
        <span className="mb-1.5 block text-sm font-bold text-ink">Nilai (Rp)</span>
        <input
          required
          type="number"
          min={1}
          placeholder="Contoh: 5000000"
          className="w-full rounded-2xl border border-brand-100 bg-white px-4 py-3 text-sm font-semibold focus:outline-none focus:ring-2 focus:ring-brand-500"
          value={form.nilai}
          onChange={(e) => setForm({ ...form, nilai: e.target.value })}
        />
      </label>

      <label className="block">
        <span className="mb-1.5 block text-sm font-bold text-ink">Tanggal</span>
        <input
          required
          type="date"
          max={new Date().toISOString().slice(0, 10)}
          className="w-full rounded-2xl border border-brand-100 bg-white px-4 py-3 text-sm font-semibold focus:outline-none focus:ring-2 focus:ring-brand-500"
          value={form.tanggal}
          onChange={(e) => setForm({ ...form, tanggal: e.target.value })}
        />
      </label>

      <button
        type="submit"
        disabled={loading}
        className="flex items-center gap-2 rounded-full bg-brand-600 px-6 py-3 text-sm font-black text-white shadow hover:bg-brand-700 disabled:opacity-50 transition"
      >
        <Banknote className="h-4 w-4" />
        {loading ? "Menyimpan..." : "Kirim TYFCB"}
      </button>
    </form>
  );
}

// ─── Visitor Form ────────────────────────────────────────────────────────────

function VisitorForm({ members }: { members: TeamMember[] }) {
  const router = useRouter();
  const [form, setForm] = useState({ member_id: "", nama: "", kontak: "", tanggal_undang: "" });
  const [loading, setLoading] = useState(false);
  const [msg, setMsg] = useState<{ ok: boolean; text: string } | null>(null);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setMsg(null);
    setLoading(true);

    const res = await fetch("/api/captain/visitors", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(form),
    });

    const data = await res.json();
    setLoading(false);

    if (!res.ok) {
      setMsg({ ok: false, text: data.error ?? "Gagal menyimpan." });
      return;
    }

    setMsg({ ok: true, text: "Visitor berhasil didaftarkan." });
    setForm({ member_id: "", nama: "", kontak: "", tanggal_undang: "" });
    router.refresh();
  }

  return (
    <form onSubmit={submit} className="space-y-4 max-w-lg">
      {msg && (
        <div className={`rounded-2xl border px-4 py-3 text-sm font-semibold ${
          msg.ok ? "border-green-200 bg-green-50 text-green-700" : "border-brand-100 bg-brand-50 text-brand-700"
        }`}>
          {msg.text}
        </div>
      )}

      <label className="block">
        <span className="mb-1.5 block text-sm font-bold text-ink">Anggota pengundang</span>
        <select
          required
          className="w-full rounded-2xl border border-brand-100 bg-white px-4 py-3 text-sm font-semibold focus:outline-none focus:ring-2 focus:ring-brand-500"
          value={form.member_id}
          onChange={(e) => setForm({ ...form, member_id: e.target.value })}
        >
          <option value="">— Pilih anggota</option>
          {members.map((m) => (
            <option key={m.id} value={m.id}>{m.full_name}</option>
          ))}
        </select>
      </label>

      <label className="block">
        <span className="mb-1.5 block text-sm font-bold text-ink">Nama visitor</span>
        <input
          required
          type="text"
          placeholder="Nama lengkap visitor"
          className="w-full rounded-2xl border border-brand-100 bg-white px-4 py-3 text-sm font-semibold focus:outline-none focus:ring-2 focus:ring-brand-500"
          value={form.nama}
          onChange={(e) => setForm({ ...form, nama: e.target.value })}
        />
      </label>

      <label className="block">
        <span className="mb-1.5 block text-sm font-bold text-ink">Kontak (WhatsApp)</span>
        <input
          required
          type="text"
          placeholder="08xxxxxxxxxx"
          className="w-full rounded-2xl border border-brand-100 bg-white px-4 py-3 text-sm font-semibold focus:outline-none focus:ring-2 focus:ring-brand-500"
          value={form.kontak}
          onChange={(e) => setForm({ ...form, kontak: e.target.value })}
        />
      </label>

      <label className="block">
        <span className="mb-1.5 block text-sm font-bold text-ink">Tanggal undang</span>
        <input
          required
          type="date"
          max={new Date().toISOString().slice(0, 10)}
          className="w-full rounded-2xl border border-brand-100 bg-white px-4 py-3 text-sm font-semibold focus:outline-none focus:ring-2 focus:ring-brand-500"
          value={form.tanggal_undang}
          onChange={(e) => setForm({ ...form, tanggal_undang: e.target.value })}
        />
      </label>

      <button
        type="submit"
        disabled={loading}
        className="flex items-center gap-2 rounded-full bg-brand-600 px-6 py-3 text-sm font-black text-white shadow hover:bg-brand-700 disabled:opacity-50 transition"
      >
        <UserPlus className="h-4 w-4" />
        {loading ? "Menyimpan..." : "Daftarkan Visitor"}
      </button>
    </form>
  );
}

// ─── Password Reset ──────────────────────────────────────────────────────────

function PasswordResetInline({ memberId }: { memberId: string }) {
  const [open, setOpen] = useState(false);
  const [pw, setPw] = useState("");
  const [show, setShow] = useState(false);
  const [loading, setLoading] = useState(false);
  const [msg, setMsg] = useState<{ ok: boolean; text: string } | null>(null);

  async function save() {
    if (pw.length < 6) {
      setMsg({ ok: false, text: "Password minimal 6 karakter." });
      return;
    }
    setLoading(true);
    const res = await fetch(`/api/captain/members/${memberId}/password`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ new_password: pw }),
    });
    const data = await res.json();
    setLoading(false);
    if (!res.ok) {
      setMsg({ ok: false, text: data.error ?? "Gagal." });
      return;
    }
    setMsg({ ok: true, text: "Password berhasil diubah." });
    setPw("");
    setOpen(false);
  }

  if (!open) {
    return (
      <button
        onClick={() => { setOpen(true); setMsg(null); }}
        className="flex items-center gap-1.5 rounded-lg border border-brand-100 bg-white px-3 py-1.5 text-xs font-bold text-muted hover:border-brand-300 hover:text-brand-700 transition"
      >
        <KeyRound className="h-3.5 w-3.5" />
        Reset Password
      </button>
    );
  }

  return (
    <div className="mt-2 flex flex-col gap-2">
      {msg && (
        <p className={`text-xs font-semibold ${msg.ok ? "text-green-700" : "text-brand-700"}`}>{msg.text}</p>
      )}
      <div className="flex items-center gap-2">
        <div className="relative flex-1">
          <input
            type={show ? "text" : "password"}
            placeholder="Password baru (min 6 char)"
            className="w-full rounded-xl border border-brand-200 bg-white px-3 py-2 pr-9 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500"
            value={pw}
            onChange={(e) => setPw(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && save()}
            autoFocus
          />
          <button
            type="button"
            onClick={() => setShow((v) => !v)}
            className="absolute right-2.5 top-2 text-muted"
          >
            {show ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
          </button>
        </div>
        <button
          onClick={save}
          disabled={loading}
          className="flex items-center gap-1 rounded-xl bg-brand-600 px-3 py-2 text-xs font-bold text-white hover:bg-brand-700 disabled:opacity-50 transition"
        >
          <Check className="h-3.5 w-3.5" />
          {loading ? "..." : "Simpan"}
        </button>
        <button
          onClick={() => { setOpen(false); setMsg(null); setPw(""); }}
          className="rounded-xl border border-brand-100 bg-white p-2 text-muted hover:text-ink transition"
        >
          <X className="h-3.5 w-3.5" />
        </button>
      </div>
    </div>
  );
}

// ─── Void button ─────────────────────────────────────────────────────────────

function VoidButton({ url, label, onDone }: { url: string; label: string; onDone: () => void }) {
  const [loading, setLoading] = useState(false);

  async function doVoid() {
    if (!window.confirm(`Void "${label}"?`)) return;
    setLoading(true);
    const res = await fetch(url, { method: "PATCH" });
    setLoading(false);
    if (!res.ok) {
      const d = await res.json();
      alert(d.error ?? "Gagal void.");
      return;
    }
    onDone();
  }

  return (
    <button
      onClick={doVoid}
      disabled={loading}
      title="Void entry ini"
      className="flex items-center gap-1 rounded-lg border border-red-100 bg-red-50 px-2 py-1 text-xs font-bold text-red-700 hover:bg-red-100 disabled:opacity-50 transition"
    >
      <Trash2 className="h-3 w-3" />
      {loading ? "..." : "Void"}
    </button>
  );
}

// ─── Member List Tab ──────────────────────────────────────────────────────────

function MemberListTab({
  members,
  pendingTyfcb,
  terdaftarVisitors,
  onRefresh,
}: {
  members: TeamMember[];
  pendingTyfcb: PendingTyfcb[];
  terdaftarVisitors: TerdaftarVisitor[];
  onRefresh: () => void;
}) {
  return (
    <div className="space-y-4">
      {members.map((m) => {
        const myTyfcb  = pendingTyfcb.filter((t) => t.seller_id === m.id);
        const myVisitors = terdaftarVisitors.filter((v) => v.inviter_id === m.id);

        return (
          <div key={m.id} className="glass-panel overflow-hidden rounded-2xl">
            {/* Member header */}
            <div className="flex flex-wrap items-center justify-between gap-3 border-b border-brand-50 px-4 py-3">
              <div>
                <p className="font-black text-ink">{m.full_name}</p>
                <p className="text-xs text-muted">{m.email}</p>
              </div>
              <div className="flex items-center gap-2">
                <span className={`rounded-full px-2.5 py-0.5 text-xs font-bold capitalize ${COLOR_BADGE[m.color_status] ?? "bg-slate-50 text-slate-500"}`}>
                  {m.color_status}
                </span>
                <PasswordResetInline memberId={m.id} />
              </div>
            </div>

            <div className="divide-y divide-brand-50">
              {/* Pending TYFCB */}
              {myTyfcb.length > 0 && (
                <div className="px-4 py-3">
                  <p className="mb-2 text-xs font-bold uppercase tracking-wide text-muted">
                    TYFCB Pending ({myTyfcb.length})
                  </p>
                  <div className="space-y-2">
                    {myTyfcb.map((t) => (
                      <div key={t.id} className="flex items-center justify-between gap-3 rounded-xl border border-brand-50 bg-brand-50/50 px-3 py-2">
                        <div className="min-w-0">
                          <p className="truncate text-sm font-semibold text-ink">Pembeli: {t.buyer_name}</p>
                          <p className="text-xs text-muted">
                            {formatCurrency(t.nilai)} · {t.tanggal} · {t.computed_score} poin
                          </p>
                        </div>
                        <VoidButton
                          url={`/api/captain/tyfcb/${t.id}/void`}
                          label={`TYFCB dari ${t.buyer_name}`}
                          onDone={onRefresh}
                        />
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Terdaftar Visitors */}
              {myVisitors.length > 0 && (
                <div className="px-4 py-3">
                  <p className="mb-2 text-xs font-bold uppercase tracking-wide text-muted">
                    Visitor Terdaftar ({myVisitors.length})
                  </p>
                  <div className="space-y-2">
                    {myVisitors.map((v) => (
                      <div key={v.id} className="flex items-center justify-between gap-3 rounded-xl border border-brand-50 bg-brand-50/50 px-3 py-2">
                        <div className="min-w-0">
                          <p className="truncate text-sm font-semibold text-ink">{v.nama}</p>
                          <p className="text-xs text-muted">{v.kontak} · {v.tanggal_undang}</p>
                        </div>
                        <VoidButton
                          url={`/api/captain/visitors/${v.id}/void`}
                          label={v.nama}
                          onDone={onRefresh}
                        />
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {myTyfcb.length === 0 && myVisitors.length === 0 && (
                <p className="px-4 py-3 text-xs text-muted">Tidak ada entri pending.</p>
              )}
            </div>
          </div>
        );
      })}
    </div>
  );
}

// ─── Main Panel ───────────────────────────────────────────────────────────────

export function CaptainPanel({
  members,
  allMembers,
  pendingTyfcb,
  terdaftarVisitors,
}: {
  members: TeamMember[];
  allMembers: SeasonMember[];
  pendingTyfcb: PendingTyfcb[];
  terdaftarVisitors: TerdaftarVisitor[];
}) {
  const router = useRouter();
  const [tab, setTab] = useState<TabKey>("tyfcb");

  return (
    <div>
      {/* Tab bar */}
      <div className="mb-6 flex gap-2 border-b border-brand-100 pb-1">
        {TABS.map(({ key, label, icon: Icon }) => (
          <button
            key={key}
            onClick={() => setTab(key)}
            className={`flex items-center gap-2 rounded-t-xl px-4 py-2.5 text-sm font-bold transition ${
              tab === key
                ? "border-b-2 border-brand-600 text-brand-700"
                : "text-muted hover:text-ink"
            }`}
          >
            <Icon className="h-4 w-4" />
            {label}
          </button>
        ))}
      </div>

      {tab === "tyfcb" && (
        <TyfcbForm
          members={members}
          allMembers={allMembers}
        />
      )}

      {tab === "visitor" && (
        <VisitorForm members={members} />
      )}

      {tab === "members" && (
        <MemberListTab
          members={members}
          pendingTyfcb={pendingTyfcb}
          terdaftarVisitors={terdaftarVisitors}
          onRefresh={() => router.refresh()}
        />
      )}
    </div>
  );
}
