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

	db, err := postgres.Connect(ctx, cfg.DatabaseURL)
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

	// Use cases (application layer).
	authUC := usecase.NewAuth(users, sessions)
	memberUC := usecase.NewMember(members, users, teams, classes, seasons, db)
	tyfcbUC := usecase.NewTyfcb(tyfcbRepo, members, ledger, seasons, db)
	visitorUC := usecase.NewVisitor(visitorRepo, members, ledger, db)
	catalogUC := usecase.NewCatalog(teams, classes, boosters, seasons)
	leaderboardUC := usecase.NewLeaderboard(ledger, members, tyfcbRepo, visitorRepo, boosters, badges, seasons)

	server := delivery.NewServer(delivery.Deps{
		Config:      cfg,
		DB:          db,
		Auth:        authUC,
		Members:     memberUC,
		Tyfcb:       tyfcbUC,
		Visitors:    visitorUC,
		Catalog:     catalogUC,
		Leaderboard: leaderboardUC,
		Seasons:     seasons,
		MemberRepo:  members,
	})

	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           server.Router(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       90 * time.Second,
	}

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
