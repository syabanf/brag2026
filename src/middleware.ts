import { NextRequest, NextResponse } from "next/server";

const SESSION_COOKIE = "brag_session";
const DEMO_COOKIE = "brag_demo";

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  if (
    pathname.startsWith("/_next/") ||
    pathname.startsWith("/favicon") ||
    pathname.startsWith("/welcome") ||
    pathname.startsWith("/login") ||
    pathname.startsWith("/public/") ||
    pathname.startsWith("/api/")
  ) {
    return NextResponse.next();
  }

  // Demo visitors carry no session; the demo cookie is their entry ticket.
  if (request.cookies.get(DEMO_COOKIE)?.value === "1") {
    return NextResponse.next();
  }

  if (!request.cookies.get(SESSION_COOKIE)) {
    const welcomeUrl = request.nextUrl.clone();
    welcomeUrl.pathname = "/welcome";
    welcomeUrl.search = "";
    return NextResponse.redirect(welcomeUrl);
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon\\.ico).*)"],
};
