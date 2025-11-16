package server

import (
	"net/http"

	"pr_reviewer_assignment_service/internal/http/handlers"
	"pr_reviewer_assignment_service/internal/http/middleware"
)

// registerRoutes регистрирует все маршруты для сервера.
func (s *Server) registerRoutes() {

	logMiddleware := middleware.LoggingMiddleware(s.logger)

	userHandler := handlers.NewUserHandler(s.userService)
	prHandler := handlers.NewPRHandler(s.prService)
	teamHandler := handlers.NewTeamHandler(s.teamService)

	s.mux.Handle("/users/set-active", logMiddleware(http.HandlerFunc(userHandler.SetActive)))
	s.mux.Handle("/users/get-review", logMiddleware(http.HandlerFunc(userHandler.GetReview)))

	s.mux.Handle("/pull-request/create", logMiddleware(http.HandlerFunc(prHandler.CreatePR)))
	s.mux.Handle("/pull-request/merge", logMiddleware(http.HandlerFunc(prHandler.MergePR)))
	s.mux.Handle("/pull-request/reassign", logMiddleware(http.HandlerFunc(prHandler.ReassignPR)))

	s.mux.Handle("/team/add", logMiddleware(http.HandlerFunc(teamHandler.CreateTeam)))
	s.mux.Handle("/team/get", logMiddleware(http.HandlerFunc(teamHandler.GetTeam)))
}
