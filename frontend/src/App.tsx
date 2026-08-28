import type { ReactNode } from "react";
import { BrowserRouter, Navigate, Route, Routes, useLocation } from "react-router-dom";
import { AuthProvider } from "./lib/auth-context";
import { useAuth } from "./lib/use-auth";
import { AppShell } from "./components/app-shell";
import { Spinner } from "./components/ui";
import { LoginPage } from "./pages/login";
import { DashboardPage } from "./pages/dashboard";
import { LeaderboardPage } from "./pages/leaderboard";
import { SubmitPage } from "./pages/submit";
import { CaptainPage } from "./pages/captain";
import { PrizesPage } from "./pages/prizes";
import { AdminEventsPage } from "./pages/admin-events";
import { AdminHomePage } from "./pages/admin-home";
import { AdminBoosterPage } from "./pages/admin-booster";
import { ActivityPage } from "./pages/activity";
import {
  AdminClassificationsPage,
  AdminMembersPage,
  AdminTeamsPage,
  AdminTyfcbPage,
  AdminVisitorsPage,
} from "./pages/admin";
import {
  AwardsPage,
  BoosterDetailPage,
  BoosterPage,
  HistoryPage,
  NotFoundPage,
  ProfilePage,
} from "./pages/misc";

/** Gates a route behind a session, and optionally behind a role. */
function Protected({ children, role }: { children: ReactNode; role?: "admin" | "captain" }) {
  const { user, loading } = useAuth();
  const location = useLocation();

  if (loading) return <Spinner />;
  if (!user) return <Navigate to="/login" state={{ from: location }} replace />;

  if (role === "admin" && user.role !== "admin") return <Navigate to="/" replace />;
  if (role === "captain" && user.role !== "captain" && user.role !== "admin") {
    return <Navigate to="/" replace />;
  }

  return <AppShell>{children}</AppShell>;
}

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<LoginPage />} />

          {/* Shareable board: no session required. */}
          <Route
            path="/public/leaderboard"
            element={
              <main className="mx-auto w-full max-w-4xl px-4 py-8 sm:px-6">
                <LeaderboardPage isPublic />
              </main>
            }
          />

          <Route path="/" element={<Protected><DashboardPage /></Protected>} />
          <Route path="/leaderboard" element={<Protected><LeaderboardPage /></Protected>} />
          <Route path="/submit" element={<Protected><SubmitPage /></Protected>} />
          <Route path="/booster" element={<Protected><BoosterPage /></Protected>} />
          <Route path="/booster/:id" element={<Protected><BoosterDetailPage /></Protected>} />
          <Route path="/awards" element={<Protected><AwardsPage /></Protected>} />
          <Route path="/prizes" element={<Protected><PrizesPage /></Protected>} />
          <Route path="/activity" element={<Protected><ActivityPage /></Protected>} />
          <Route path="/history" element={<Protected><HistoryPage /></Protected>} />
          <Route path="/profile" element={<Protected><ProfilePage /></Protected>} />

          <Route path="/captain" element={<Protected role="captain"><CaptainPage /></Protected>} />

          <Route path="/admin" element={<Protected role="admin"><AdminHomePage /></Protected>} />
          <Route path="/admin/booster" element={<Protected role="admin"><AdminBoosterPage /></Protected>} />
          <Route path="/admin/members" element={<Protected role="admin"><AdminMembersPage /></Protected>} />
          <Route path="/admin/teams" element={<Protected role="admin"><AdminTeamsPage /></Protected>} />
          <Route path="/admin/classifications" element={<Protected role="admin"><AdminClassificationsPage /></Protected>} />
          <Route path="/admin/tyfcb" element={<Protected role="admin"><AdminTyfcbPage /></Protected>} />
          <Route path="/admin/visitors" element={<Protected role="admin"><AdminVisitorsPage /></Protected>} />
          <Route path="/admin/events" element={<Protected role="admin"><AdminEventsPage /></Protected>} />

          <Route path="*" element={<Protected><NotFoundPage /></Protected>} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}
