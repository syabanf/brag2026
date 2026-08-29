package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"

	"github.com/ikurniawann/brag2026/backend/internal/config"
	"github.com/ikurniawann/brag2026/backend/internal/domain"
	"github.com/ikurniawann/brag2026/backend/internal/repository/postgres"
	"github.com/ikurniawann/brag2026/backend/internal/usecase"
)

// Server wires the use cases to routes. It holds no business logic itself.
type Server struct {
	cfg *config.Config
	db  *postgres.DB

	auth        *usecase.Auth
	members     *usecase.Member
	tyfcb       *usecase.Tyfcb
	visitors    *usecase.Visitor
	catalog     *usecase.Catalog
	leaderboard *usecase.Leaderboard
	passes      *usecase.ScoringPass
	prizes      *usecase.Prize
	network     *usecase.Network
	reports     *usecase.Reports

	seasons    domain.SeasonRepository
	memberRepo domain.MemberRepository
	events     domain.WeeklyEventRepository
}

type Deps struct {
	Config      *config.Config
	DB          *postgres.DB
	Auth        *usecase.Auth
	Members     *usecase.Member
	Tyfcb       *usecase.Tyfcb
	Visitors    *usecase.Visitor
	Catalog     *usecase.Catalog
	Leaderboard *usecase.Leaderboard
	Passes      *usecase.ScoringPass
	Prizes      *usecase.Prize
	Network     *usecase.Network
	Reports     *usecase.Reports
	Seasons     domain.SeasonRepository
	MemberRepo  domain.MemberRepository
	Events      domain.WeeklyEventRepository
}

func NewServer(d Deps) *Server {
	return &Server{
		cfg: d.Config, db: d.DB,
		auth: d.Auth, members: d.Members, tyfcb: d.Tyfcb,
		visitors: d.Visitors, catalog: d.Catalog, leaderboard: d.Leaderboard,
		passes: d.Passes, prizes: d.Prizes, network: d.Network, reports: d.Reports,
		seasons: d.Seasons, memberRepo: d.MemberRepo, events: d.Events,
	}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(secureHeaders)
	r.Use(limitBody)

	// Credentials are cookie-based, so the origin list must be explicit.
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: s.cfg.AllowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Content-Type", "X-Requested-With"},
		// Without this the browser hides the header on a cross-origin response,
		// and a downloaded export loses the filename the server chose for it.
		ExposedHeaders:   []string{"Content-Disposition"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(s.authenticate)

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", s.handleHealth)

		// Public: the shareable leaderboard needs no session.
		r.Group(func(r chi.Router) {
			// Ten attempts per minute per IP: enough for a fumbled password,
			// far too slow to grind through a password list.
			r.With(httprate.LimitByIP(10, time.Minute)).
				Post("/auth/login", s.handleLogin)
			r.Post("/auth/logout", s.handleLogout)
			r.Get("/public/leaderboard", s.handleLeaderboard)
			r.Get("/public/leaderboard/teams/{id}/history", s.handleTeamHistory)

			// The tour introduces the product, so it works before signing in.
			r.Get("/tour/steps", s.handleTourSteps)
			r.Get("/tour/voice", s.handleTourVoice)
		})

		// Any signed-in member.
		r.Group(func(r chi.Router) {
			r.Use(requireAuth)

			r.Get("/auth/me", s.handleMe)
			r.Post("/auth/change-password", s.handleChangePassword)

			r.Get("/dashboard", s.handleDashboard)
			r.Get("/leaderboard", s.handleLeaderboard)
			r.Get("/leaderboard/teams/{id}/history", s.handleTeamHistory)
			r.Get("/members/search", s.handleSearchMembers)
			r.Get("/boosters", s.handleListBoosters)
			r.Get("/boosters/{id}", s.handleGetBooster)
			r.Get("/badges", s.handleBadges)
			r.Get("/activity", s.handleActivity)
			r.Get("/events", s.handleListWeeklyEvents)
			r.Get("/events/bank", s.handleEventBank)
			r.Get("/prizes", s.handleListPrizes)
			r.Get("/raffle/tickets", s.handleTicketStandings)
			// Members may take the leaderboard away with them; the other
			// reports carry personal data and stay under /admin.
			r.Get("/export/leaderboard", s.handleExportLeaderboard)
			r.Get("/spheres", s.handleListSpheres)
			r.Get("/one-to-one", s.handleListOneToOne)

			// Recording a contribution needs a competition profile.
			r.Group(func(r chi.Router) {
				r.Use(s.withMember)
				r.Post("/tyfcb", s.handleSubmitTyfcb)
				r.Post("/visitors", s.handleRegisterVisitor)
				r.Post("/prizes/donate", s.handleDonatePrize)
				r.Post("/one-to-one", s.handleLogOneToOne)
			})
		})

		// Captains (and admins).
		r.Route("/captain", func(r chi.Router) {
			r.Use(requireAuth, requireCaptain)

			r.Get("/team", s.handleCaptainTeam)
			r.Post("/tyfcb", s.handleCaptainSubmitTyfcb)
			r.Patch("/tyfcb/{id}/void", s.handleVoidTyfcb)
			r.Post("/visitors", s.handleCaptainRegisterVisitor)
			r.Patch("/visitors/{id}/void", s.handleVoidVisitor)
			r.Patch("/members/{id}/password", s.handleCaptainSetPassword)
		})

		// Admin-only master data and verification.
		r.Route("/admin", func(r chi.Router) {
			r.Use(requireAuth, requireAdmin)

			r.Get("/members", s.handleListMembers)
			r.Post("/members", s.handleCreateMember)
			r.Patch("/members/{id}", s.handleUpdateMember)

			r.Get("/teams", s.handleListTeams)
			r.Post("/teams", s.handleCreateTeam)
			r.Patch("/teams/{id}", s.handleUpdateTeam)
			r.Delete("/teams/{id}", s.handleDeleteTeam)
			r.Get("/teams-meta", s.handleTeamsMeta)

			r.Get("/classifications", s.handleListClassifications)
			r.Post("/classifications", s.handleCreateClassification)
			r.Patch("/classifications/{id}", s.handleUpdateClassification)
			r.Delete("/classifications/{id}", s.handleDeleteClassification)

			r.Get("/booster", s.handleListBoosters)
			r.Post("/booster", s.handleCreateBooster)
			r.Patch("/booster/{id}", s.handleUpdateBooster)
			r.Delete("/booster/{id}", s.handleDeleteBooster)

			r.Get("/tyfcb", s.handleListTyfcb)
			r.Patch("/tyfcb/{id}", s.handleSetTyfcbStatus)

			r.Get("/visitors", s.handleListVisitors)
			r.Patch("/visitors/{id}", s.handleUpdateVisitor)

			r.Get("/events", s.handleListWeeklyEvents)
			r.Post("/events", s.handleScheduleEvent)
			r.Delete("/events/{id}", s.handleDeleteWeeklyEvent)

			// Periodic settlements. Both are idempotent per period, so the
			// committee can re-run them safely.
			r.Get("/export/{report}", s.handleExport)

			r.Post("/passes/weekly", s.handleWeeklyPass)
			r.Post("/passes/daily", s.handleDailyPass)

			r.Get("/prizes", s.handleListPrizes)
			r.Post("/prizes", s.handleSeedPrize)
			r.Patch("/prizes/{id}", s.handleSetPrizeStatus)
			r.Delete("/prizes/{id}", s.handleDeletePrize)
			r.Post("/raffle/issue", s.handleIssueTickets)
			r.Post("/raffle/draw/{id}", s.handleDrawPrize)

			r.Get("/spheres", s.handleListSpheres)
			r.Post("/spheres", s.handleCreateSphere)
			r.Patch("/spheres/{id}/members", s.handleSetSphereMembers)
			r.Delete("/spheres/{id}", s.handleDeleteSphere)
			r.Delete("/one-to-one/{id}", s.handleDeleteOneToOne)
		})
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "Endpoint tidak ditemukan.")
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusMethodNotAllowed, "Metode tidak diizinkan.")
	})

	return r
}
