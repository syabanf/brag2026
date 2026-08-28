package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

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

	seasons    domain.SeasonRepository
	memberRepo domain.MemberRepository
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
	Seasons     domain.SeasonRepository
	MemberRepo  domain.MemberRepository
}

func NewServer(d Deps) *Server {
	return &Server{
		cfg: d.Config, db: d.DB,
		auth: d.Auth, members: d.Members, tyfcb: d.Tyfcb,
		visitors: d.Visitors, catalog: d.Catalog, leaderboard: d.Leaderboard,
		seasons: d.Seasons, memberRepo: d.MemberRepo,
	}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Credentials are cookie-based, so the origin list must be explicit.
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   s.cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "X-Requested-With"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(s.authenticate)

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", s.handleHealth)

		// Public: the shareable leaderboard needs no session.
		r.Group(func(r chi.Router) {
			r.Post("/auth/login", s.handleLogin)
			r.Post("/auth/logout", s.handleLogout)
			r.Get("/public/leaderboard", s.handleLeaderboard)
			r.Get("/public/leaderboard/teams/{id}/history", s.handleTeamHistory)
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

			// Recording a contribution needs a competition profile.
			r.Group(func(r chi.Router) {
				r.Use(s.withMember)
				r.Post("/tyfcb", s.handleSubmitTyfcb)
				r.Post("/visitors", s.handleRegisterVisitor)
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
