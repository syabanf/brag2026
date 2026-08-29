// Command api is the composition root: it is the only place that knows every
// concrete type, wiring repositories into use cases into HTTP handlers.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/ikurniawann/brag2026/backend/internal/config"
	delivery "github.com/ikurniawann/brag2026/backend/internal/delivery/http"
	"github.com/ikurniawann/brag2026/backend/internal/repository/postgres"
	"github.com/ikurniawann/brag2026/backend/internal/usecase"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	// A missing .env is fine: real environment variables are the source of
	// truth in containers.
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("configuration", "err", err)
		os.Exit(1)
	}

	ctx := context.Background()

	db, err := postgres.Connect(ctx, cfg.DatabaseURL, cfg.DBMaxConns)
	if err != nil {
		slog.Error("database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	// Repositories (outermost layer, implementing domain contracts).
	users := postgres.NewUserRepo(db)
	sessions := postgres.NewSessionRepo(db)
	seasons := postgres.NewSeasonRepo(db)
	members := postgres.NewMemberRepo(db)
	teams := postgres.NewTeamRepo(db)
	classes := postgres.NewClassificationRepo(db)
	tyfcbRepo := postgres.NewTyfcbRepo(db)
	visitorRepo := postgres.NewVisitorRepo(db)
	ledger := postgres.NewLedgerRepo(db)
	boosters := postgres.NewBoosterRepo(db)
	badges := postgres.NewBadgeRepo(db)
	events := postgres.NewWeeklyEventRepo(db)
	passRepo := postgres.NewScoringPassRepo(db)
	prizeRepo := postgres.NewPrizeRepo(db)
	activityRepo := postgres.NewActivityRepo(db)
	sphereRepo := postgres.NewContactSphereRepo(db)
	oneToOneRepo := postgres.NewOneToOneRepo(db)

	// Use cases (application layer).
	badgeUC := usecase.NewBadges(badges)
	authUC := usecase.NewAuth(users, sessions)
	memberUC := usecase.NewMember(members, users, teams, classes, seasons, ledger, badgeUC, db)
	tyfcbUC := usecase.NewTyfcb(tyfcbRepo, members, ledger, seasons, events, sphereRepo, badgeUC, db)
	visitorUC := usecase.NewVisitor(visitorRepo, members, ledger, events, badgeUC, db)
	catalogUC := usecase.NewCatalog(teams, classes, boosters, seasons)
	passUC := usecase.NewScoringPass(passRepo, ledger, events, oneToOneRepo, seasons, badgeUC, db)
	prizeUC := usecase.NewPrize(prizeRepo, members, seasons, badgeUC, db)
	networkUC := usecase.NewNetwork(sphereRepo, oneToOneRepo, members, seasons, db)
	reportsUC := usecase.NewReports(ledger, members, tyfcbRepo, visitorRepo, prizeRepo, seasons)
	leaderboardUC := usecase.NewLeaderboard(ledger, members, tyfcbRepo, visitorRepo, boosters, badges, activityRepo, seasons)

	server := delivery.NewServer(delivery.Deps{
		Config:      cfg,
		DB:          db,
		Auth:        authUC,
		Members:     memberUC,
		Tyfcb:       tyfcbUC,
		Visitors:    visitorUC,
		Catalog:     catalogUC,
		Leaderboard: leaderboardUC,
		Passes:      passUC,
		Prizes:      prizeUC,
		Network:     networkUC,
		Reports:     reportsUC,
		Seasons:     seasons,
		MemberRepo:  members,
		Events:      events,
	})

	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           server.Router(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       90 * time.Second,
	}

	// Expired sessions are dead weight: nothing can use them, but they keep
	// the token index growing for the life of the season.
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()

		for {
			removed, err := sessions.DeleteExpired(context.Background())
			if err != nil {
				slog.Error("session sweep", "err", err)
			} else if removed > 0 {
				slog.Info("swept expired sessions", "removed", removed)
			}
			<-ticker.C
		}
	}()

	go func() {
		slog.Info("listening", "port", cfg.Port, "origins", cfg.AllowedOrigins)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server", "err", err)
			os.Exit(1)
		}
	}()

	// Drain in-flight requests before exiting so a deploy never cuts a
	// verification mid-transaction.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownGrace)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown", "err", err)
	}
}
