package handlers

import (
	"encoding/json"
	"net/http"
	"pr_reviewer_assignment_service/internal/dto"
	"pr_reviewer_assignment_service/internal/dto/team"
	usecase "pr_reviewer_assignment_service/internal/usecase/team"

	"go.uber.org/zap"
)

type TeamHandler struct {
	svc *usecase.TeamService
}

func NewTeamHandler(svc *usecase.TeamService) *TeamHandler {
	return &TeamHandler{svc: svc}
}

func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req team.TeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.svc.Logger().Error(ctx, "failed to decode request", zap.Error(err))
		writeError(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	h.svc.Logger().Info(ctx, "CreateTeam request received", zap.String("team_name", req.TeamName))

	resp, err := h.svc.CreateTeam(ctx, &req)
	if err != nil {
		h.svc.Logger().Error(ctx, "CreateTeam failed", zap.Error(err), zap.String("team_name", req.TeamName))
		switch err {
		case dto.ErrTeamExists:
			writeError(w, http.StatusConflict, "TEAM_EXISTS", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		}
		return
	}

	h.svc.Logger().Info(ctx, "CreateTeam succeeded", zap.String("team_name", req.TeamName))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		h.svc.Logger().Error(ctx, "GetTeam missing team_name")
		writeError(w, http.StatusBadRequest, "INVALID_INPUT", "team_name required")
		return
	}

	h.svc.Logger().Info(ctx, "GetTeam request received", zap.String("team_name", teamName))

	resp, err := h.svc.GetTeamByName(ctx, teamName)
	if err != nil {
		h.svc.Logger().Error(ctx, "GetTeam failed", zap.Error(err), zap.String("team_name", teamName))
		switch err {
		case dto.ErrNotFound:
			writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		}
		return
	}

	h.svc.Logger().Info(ctx, "GetTeam succeeded", zap.String("team_name", teamName))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
