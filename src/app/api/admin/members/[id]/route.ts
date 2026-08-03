import { NextRequest } from "next/server";
import { query } from "@/lib/db";
import { getCurrentUser } from "@/lib/local-auth";

export async function PATCH(
  req: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const user = await getCurrentUser();
  if (!user) return Response.json({ error: "Unauthorized" }, { status: 401 });
  if (user.role !== "admin") return Response.json({ error: "Forbidden" }, { status: 403 });

  const { id } = await params;
  const body = await req.json();
  const { full_name, email, team_id, klasifikasi_id, color_status, is_active, role, new_password } = body;

  if (new_password !== undefined && new_password !== "") {
    if (typeof new_password !== "string" || new_password.length < 6) {
      return Response.json({ error: "Password baru minimal 6 karakter." }, { status: 400 });
    }
  }

  // Role promotion/demotion — separate path, cannot demote yourself
  if (role !== undefined) {
    if (role !== "admin" && role !== "captain" && role !== "member") {
      return Response.json({ error: "Role tidak valid." }, { status: 400 });
    }

    const { rows: target } = await query<{ user_id: string }>(
      `select user_id from members where id = $1`, [id]
    );
    if (!target[0]) return Response.json({ error: "Member tidak ditemukan." }, { status: 404 });
    if (role !== "admin" && target[0].user_id === user.id) {
      return Response.json({ error: "Tidak bisa menurunkan role diri sendiri." }, { status: 400 });
    }

    await query(`update app_users set role = $1 where id = $2`, [role, target[0].user_id]);
    return Response.json({ ok: true });
  }

  // Update competition profile
  await query(
    `update members
     set team_id = $1, klasifikasi_id = $2, color_status = $3, is_active = $4
     where id = $5`,
    [team_id || null, klasifikasi_id || null, color_status, is_active ?? true, id]
  );

  // Update user account info if provided
  if (full_name || email) {
    await query(
      `update app_users u
       set full_name = coalesce($1, u.full_name),
           email     = coalesce($2, u.email)
       from members m
       where m.id = $3 and u.id = m.user_id`,
      [full_name || null, email || null, id]
    );
  }

  // Admin password reset — hashed with pgcrypto, sessions invalidated
  if (new_password) {
    await query(
      `update app_users u
       set password_hash = crypt($1, gen_salt('bf'))
       from members m
       where m.id = $2 and u.id = m.user_id`,
      [new_password, id]
    );
    await query(
      `delete from user_sessions s
       using members m
       where m.id = $1 and s.user_id = m.user_id and m.user_id <> $2`,
      [id, user.id]
    );
  }

  return Response.json({ ok: true });
}
