import { NextResponse } from "next/server";
import { getTeamHistory, UUID_PATTERN } from "@/lib/domain/team-history";

// Unauthenticated on purpose — it backs /public/leaderboard, which is shared
// as a link with people who have no BRAG account. It exposes exactly what the
// public leaderboard already implies (team standings and their contributors)
// and nothing more; visitor contact details are never selected.
export async function GET(
  _req: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  if (!UUID_PATTERN.test(id)) {
    return NextResponse.json({ error: "Team tidak valid." }, { status: 400 });
  }

  const history = await getTeamHistory(id);
  if (!history) {
    return NextResponse.json({ error: "Team tidak ditemukan." }, { status: 404 });
  }

  return NextResponse.json(history);
}
