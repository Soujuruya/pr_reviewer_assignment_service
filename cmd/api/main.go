package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pr_reviewer_assignment_service/internal/config"
	"pr_reviewer_assignment_service/internal/repository/postgres"
	"pr_reviewer_assignment_service/internal/server"
	usecasePr "pr_reviewer_assignment_service/internal/usecase/pr"
	usecaseTeam "pr_reviewer_assignment_service/internal/usecase/team"
	usecaseUser "pr_reviewer_assignment_service/internal/usecase/user"
	"pr_reviewer_assignment_service/pkg/logger"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.ParseConfig("configs/.env")
	if err != nil {
		fmt.Printf("failed to parse config: %v\n", err)
		os.Exit(1)
	}

	log := logger.NewLogger(cfg.Environment)
	defer log.Sync()

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User,
		cfg.Postgres.Password, cfg.Postgres.DBName, cfg.Postgres.SSLMode,
	)
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Error(context.Background(), "failed to connect to postgres", zap.Error(err))
		os.Exit(1)
	}
	defer db.Close()

	userRepo := postgres.NewUserRepository(db, log)
	teamRepo := postgres.NewTeamRepository(db, log)
	prRepo := postgres.NewPRRepository(db, log)

	userSvc := usecaseUser.NewUserService(userRepo, prRepo, log)
	teamSvc := usecaseTeam.NewTeamService(teamRepo, log)
	prSvc := usecasePr.NewPRService(prRepo, teamRepo, userRepo, log)

	srv := server.NewServer(cfg, log, userSvc, prSvc, teamSvc)

	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			log.Error(context.Background(), "server error", zap.Error(err))
		}
	}()

	log.Info(context.Background(), "server started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Info(context.Background(), "shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error(ctx, "server forced to shutdown", zap.Error(err))
	} else {
		log.Info(ctx, "server exited gracefully")
	}
}
