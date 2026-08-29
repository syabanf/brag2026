import type {
  ActivityItem,
  Badge,
  BoosterEvent,
  CaptainTeam,
  Classification,
  ContactSphere,
  Dashboard,
  Leaderboard,
  LedgerEntry,
  Member,
  MemberQuery,
  OneToOne,
  Paged,
  EventBankEntry,
  PassResult,
  Prize,
  Team,
  TicketSummary,
  TourStep,
  TyfcbEntry,
  TyfcbPage,
  TyfcbQuery,
  WeeklyEvent,
  User,
  Visitor,
  VisitorQuery,
} from "./types";

const BASE = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

export class ApiError extends Error {
  status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

/**
 * Every call carries the session cookie. The backend replies with
 * `{ error }` on failure, which is unwrapped here so callers only ever deal
 * with a thrown ApiError carrying a message already written for the user.
 */
async function request<T>(path: string, init?: RequestInit): Promise<T> {
  let res: Response;
  try {
    res = await fetch(`${BASE}/api${path}`, {
      ...init,
      credentials: "include",
      headers: {
        ...(init?.body ? { "Content-Type": "application/json" } : {}),
        ...init?.headers,
      },
    });
  } catch {
    throw new ApiError("Tidak bisa terhubung ke server.", 0);
  }

  if (res.status === 204) {
    return undefined as T;
  }

  const text = await res.text();
  const payload = text ? JSON.parse(text) : null;

  if (!res.ok) {
    throw new ApiError(payload?.error ?? "Terjadi kesalahan.", res.status);
  }

  return payload as T;
}

/** Drops empty values so an unset filter never reaches the server. */
function qs(params: Record<string, string | number | undefined>) {
  const search = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== "") search.set(key, String(value));
  }
  const rendered = search.toString();
  return rendered ? `?${rendered}` : "";
}

const get = <T>(path: string) => request<T>(path);
const post = <T>(path: string, body?: unknown) =>
  request<T>(path, { method: "POST", body: body ? JSON.stringify(body) : undefined });
const patch = <T>(path: string, body?: unknown) =>
  request<T>(path, { method: "PATCH", body: body ? JSON.stringify(body) : undefined });
const del = <T>(path: string) => request<T>(path, { method: "DELETE" });

export const api = {
  auth: {
    login: (email: string, password: string) =>
      post<{ user: User }>("/auth/login", { email, password }),
    logout: () => post<{ ok: boolean }>("/auth/logout"),
    me: () => get<{ user: User; member: Member | null }>("/auth/me"),
    changePassword: (current_password: string, new_password: string) =>
      post<{ ok: boolean }>("/auth/change-password", { current_password, new_password }),
  },

  dashboard: () => get<Dashboard>("/dashboard"),

  leaderboard: {
    get: (kategori?: string) =>
      get<Leaderboard>(`/leaderboard${kategori && kategori !== "overall" ? `?kategori=${kategori}` : ""}`),
    public: (kategori?: string) =>
      get<Leaderboard>(
        `/public/leaderboard${kategori && kategori !== "overall" ? `?kategori=${kategori}` : ""}`,
      ),
    teamHistory: (teamId: string, kategori?: string) =>
      get<LedgerEntry[]>(
        `/leaderboard/teams/${teamId}/history${kategori ? `?kategori=${kategori}` : ""}`,
      ),
    publicTeamHistory: (teamId: string, kategori?: string) =>
      get<LedgerEntry[]>(
        `/public/leaderboard/teams/${teamId}/history${kategori ? `?kategori=${kategori}` : ""}`,
      ),
  },

  members: {
    search: (q: string) => get<Member[]>(`/members/search?q=${encodeURIComponent(q)}`),
  },

  tyfcb: {
    submit: (buyer_id: string, nilai: number, tanggal: string) =>
      post<TyfcbEntry>("/tyfcb", { buyer_id, nilai, tanggal }),
  },

  visitors: {
    register: (nama: string, kontak: string, tanggal_undang: string) =>
      post<Visitor>("/visitors", { nama, kontak, tanggal_undang }),
  },

  boosters: {
    list: (activeOnly = false) =>
      get<BoosterEvent[]>(`/boosters${activeOnly ? "?active=true" : ""}`),
    get: (id: string) => get<BoosterEvent>(`/boosters/${id}`),
  },

  badges: () => get<Badge[]>("/badges"),

  activity: (limit = 50) => get<ActivityItem[]>(`/activity?limit=${limit}`),

  tour: {
    steps: () => get<TourStep[]>("/tour/steps"),
    /**
     * Returns the narration audio, or null when the server has no
     * text-to-speech credentials — the caller then uses the browser voice.
     */
    voice: async (stepId: string): Promise<Blob | null> => {
      const res = await fetch(`${BASE}/api/tour/voice?step=${stepId}`, {
        credentials: "include",
      });
      const type = res.headers.get("content-type") ?? "";
      if (!res.ok || !type.startsWith("audio/")) return null;
      return res.blob();
    },
  },

  events: {
    list: () => get<WeeklyEvent[]>("/events"),
    bank: () => get<EventBankEntry[]>("/events/bank"),
  },

  prizes: {
    list: (status?: string) => get<Prize[]>(`/prizes${status ? `?status=${status}` : ""}`),
    donate: (body: Record<string, unknown>) => post<{ id: string }>("/prizes/donate", body),
  },

  raffle: {
    tickets: () => get<TicketSummary[]>("/raffle/tickets"),
  },

  spheres: () => get<ContactSphere[]>("/spheres"),

  oneToOne: {
    list: (limit = 50) => get<OneToOne[]>(`/one-to-one?limit=${limit}`),
    log: (member_id: string, tanggal: string, catatan?: string | null) =>
      post<{ id: string }>("/one-to-one", { member_id, tanggal, catatan: catatan ?? null }),
  },

  captain: {
    team: () => get<CaptainTeam>("/captain/team"),
    submitTyfcb: (member_id: string, buyer_id: string, nilai: number, tanggal: string) =>
      post<TyfcbEntry>("/captain/tyfcb", { member_id, buyer_id, nilai, tanggal }),
    voidTyfcb: (id: string) => patch<{ ok: boolean }>(`/captain/tyfcb/${id}/void`),
    registerVisitor: (member_id: string, nama: string, kontak: string, tanggal_undang: string) =>
      post<Visitor>("/captain/visitors", { member_id, nama, kontak, tanggal_undang }),
    voidVisitor: (id: string) => patch<{ ok: boolean }>(`/captain/visitors/${id}/void`),
    setPassword: (id: string, new_password: string) =>
      patch<{ ok: boolean }>(`/captain/members/${id}/password`, { new_password }),
  },

  admin: {
    members: {
      list: (query: MemberQuery = {}) => get<Paged<Member>>(`/admin/members${qs(query)}`),
      create: (body: Record<string, unknown>) => post<{ id: string }>("/admin/members", body),
      update: (id: string, body: Record<string, unknown>) =>
        patch<{ ok: boolean }>(`/admin/members/${id}`, body),
    },
    teams: {
      list: () => get<Team[]>("/admin/teams"),
      meta: () => get<{ teams: Team[]; classifications: Classification[] }>("/admin/teams-meta"),
      create: (nama_tim: string) => post<{ id: string }>("/admin/teams", { nama_tim }),
      rename: (id: string, nama_tim: string) =>
        patch<{ ok: boolean }>(`/admin/teams/${id}`, { nama_tim }),
      remove: (id: string) => del<{ ok: boolean }>(`/admin/teams/${id}`),
    },
    classifications: {
      list: () => get<Classification[]>("/admin/classifications"),
      create: (nama: string) => post<{ id: string }>("/admin/classifications", { nama }),
      rename: (id: string, nama: string) =>
        patch<{ ok: boolean }>(`/admin/classifications/${id}`, { nama }),
      remove: (id: string) => del<{ ok: boolean }>(`/admin/classifications/${id}`),
    },
    boosters: {
      list: () => get<BoosterEvent[]>("/admin/booster"),
      create: (body: Record<string, unknown>) => post<{ id: string }>("/admin/booster", body),
      update: (id: string, body: Record<string, unknown>) =>
        patch<{ ok: boolean }>(`/admin/booster/${id}`, body),
      remove: (id: string) => del<{ ok: boolean }>(`/admin/booster/${id}`),
    },
    tyfcb: {
      list: (query: TyfcbQuery = {}) => get<TyfcbPage>(`/admin/tyfcb${qs(query)}`),
      setStatus: (id: string, status: string, reason?: string) =>
        patch<{ ok: boolean }>(`/admin/tyfcb/${id}`, { status, reason: reason ?? null }),
    },
    visitors: {
      list: (query: VisitorQuery = {}) => get<Paged<Visitor>>(`/admin/visitors${qs(query)}`),
      update: (id: string, body: { status_hadir?: string; is_converted?: boolean }) =>
        patch<{ ok: boolean }>(`/admin/visitors/${id}`, body),
    },
    events: {
      list: () => get<WeeklyEvent[]>("/admin/events"),
      schedule: (body: Record<string, unknown>) => post<{ id: string }>("/admin/events", body),
      remove: (id: string) => del<{ ok: boolean }>(`/admin/events/${id}`),
    },
    passes: {
      weekly: (day?: string) =>
        post<PassResult>(`/admin/passes/weekly${day ? `?day=${day}` : ""}`),
      daily: (day?: string) =>
        post<PassResult>(`/admin/passes/daily${day ? `?day=${day}` : ""}`),
    },
    prizes: {
      list: (status?: string) => get<Prize[]>(`/admin/prizes${status ? `?status=${status}` : ""}`),
      seed: (body: Record<string, unknown>) => post<{ id: string }>("/admin/prizes", body),
      setStatus: (id: string, status: string, pemenang_id?: string | null) =>
        patch<{ ok: boolean }>(`/admin/prizes/${id}`, { status, pemenang_id: pemenang_id ?? null }),
      remove: (id: string) => del<{ ok: boolean }>(`/admin/prizes/${id}`),
    },
    raffle: {
      issue: () => post<TicketSummary[]>("/admin/raffle/issue"),
      draw: (prizeId: string) => post<Prize>(`/admin/raffle/draw/${prizeId}`),
    },
    spheres: {
      list: () => get<ContactSphere[]>("/admin/spheres"),
      create: (nama: string, deskripsi: string | null, klasifikasi_ids: string[]) =>
        post<{ id: string }>("/admin/spheres", { nama, deskripsi, klasifikasi_ids }),
      setMembers: (id: string, klasifikasi_ids: string[]) =>
        patch<{ ok: boolean }>(`/admin/spheres/${id}/members`, { klasifikasi_ids }),
      remove: (id: string) => del<{ ok: boolean }>(`/admin/spheres/${id}`),
    },
    oneToOne: {
      listAll: (limit = 100) => get<OneToOne[]>(`/one-to-one?all=true&limit=${limit}`),
      remove: (id: string) => del<{ ok: boolean }>(`/admin/one-to-one/${id}`),
    },
  },
};
