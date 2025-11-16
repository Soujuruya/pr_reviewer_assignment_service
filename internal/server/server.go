package server

import (
	"context"
	"fmt"
	"net/http"

	"pr_reviewer_assignment_service/internal/config"
	usecasePr "pr_reviewer_assignment_service/internal/usecase/pr"
	usecaseTeam "pr_reviewer_assignment_service/internal/usecase/team"
	usecaseUser "pr_reviewer_assignment_service/internal/usecase/user"

	"pr_reviewer_assignment_service/pkg/logger"
)

type Server struct {
	mux         *http.ServeMux
	logger      logger.Logger
	userService *usecaseUser.UserService
	prService   *usecasePr.PRService
	teamService *usecaseTeam.TeamService
	httpServer  *http.Server
}

func NewServer(cfg *config.Config, l logger.Logger,
	userSvc *usecaseUser.UserService,
	prSvc *usecasePr.PRService,
	teamSvc *usecaseTeam.TeamService,
) *Server {

	mux := http.NewServeMux()

	s := &Server{
		mux:         mux,
		logger:      l,
		userService: userSvc,
		prService:   prSvc,
		teamService: teamSvc,
	}

	s.registerRoutes()

	addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
	s.httpServer = &http.Server{
		Addr:           addr,
		Handler:        mux,
		ReadTimeout:    cfg.HTTP.ReadTimeout,
		WriteTimeout:   cfg.HTTP.WriteTimeout,
		IdleTimeout:    cfg.HTTP.IdleTimeout,
		MaxHeaderBytes: cfg.HTTP.MaxHeaderBytes,
	}

	return s
}

func (s *Server) Start() error {
	s.logger.Info(context.Background(), "Starting server on :8080")
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info(ctx, "Shutting down server")
	return s.httpServer.Shutdown(ctx)
}
