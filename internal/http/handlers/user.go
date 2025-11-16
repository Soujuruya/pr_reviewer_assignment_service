package handlers

import (
	"encoding/json"
	"net/http"
	"pr_reviewer_assignment_service/internal/dto"
	"pr_reviewer_assignment_service/internal/dto/user"
	usecase "pr_reviewer_assignment_service/internal/usecase/user"
	"strings"

	"go.uber.org/zap"
)

type UserHandler struct {
	svc *usecase.UserService
}

func NewUserHandler(svc *usecase.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) SetActive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req user.SetIsActiveRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.svc.Logger().Error(ctx, "failed to decode SetActive request", zap.Error(err))
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	h.svc.Logger().Info(ctx, "SetActive request received", zap.String("user_id", req.UserID), zap.Bool("is_active", req.IsActive))

	resp, err := h.svc.SetActive(ctx, &req)
	if err != nil {
		h.svc.Logger().Error(ctx, "SetActive failed", zap.Error(err), zap.String("user_id", req.UserID))
		switch err {
		case dto.ErrNotFound:
			writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
		return
	}

	h.svc.Logger().Info(ctx, "SetActive succeeded", zap.String("user_id", req.UserID), zap.Bool("is_active", req.IsActive))
	writeJSON(w, http.StatusOK, map[string]interface{}{"user": resp})
}

func (h *UserHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := strings.TrimSpace(r.URL.Query().Get("user_id"))
	if userID == "" {
		h.svc.Logger().Error(ctx, "GetReview missing user_id")
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "user_id required")
		return
	}

	h.svc.Logger().Info(ctx, "GetReview request received", zap.String("user_id", userID))

	resp, err := h.svc.GetReview(ctx, userID)
	if err != nil {
		h.svc.Logger().Error(ctx, "GetReview failed", zap.Error(err), zap.String("user_id", userID))
		switch err {
		case dto.ErrNotFound:
			writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
		return
	}

	h.svc.Logger().Info(ctx, "GetReview succeeded", zap.String("user_id", userID))
	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
