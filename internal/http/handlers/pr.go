package handlers

import (
	"encoding/json"
	"net/http"

	"pr_reviewer_assignment_service/internal/dto"
	"pr_reviewer_assignment_service/internal/dto/pr"
	usecase "pr_reviewer_assignment_service/internal/usecase/pr"

	"go.uber.org/zap"
)

type PRHandler struct {
	svc *usecase.PRService
}

func NewPRHandler(svc *usecase.PRService) *PRHandler {
	return &PRHandler{svc: svc}
}

func (h *PRHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
	var req pr.CreatePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.svc.Logger().Error(r.Context(), "Failed to decode CreatePRRequest", zap.Error(err))
		writeError(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	h.svc.Logger().Info(r.Context(), "CreatePR request received",
		zap.String("pull_request_id", req.PullRequestID),
		zap.String("author_id", req.AuthorID),
		zap.String("pull_request_name", req.PullRequestName),
	)

	resp, err := h.svc.CreatePR(r.Context(), &req)
	if err != nil {
		h.svc.Logger().Error(r.Context(), "CreatePR failed", zap.Error(err))
		switch err {
		case dto.ErrPRExists:
			writeError(w, http.StatusConflict, "PR_EXISTS", err.Error())
		case dto.ErrNotFound:
			writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		}
		return
	}

	h.svc.Logger().Info(r.Context(), "CreatePR succeeded", zap.String("pull_request_id", resp.PullRequestID))
	writeJSON(w, http.StatusCreated, map[string]interface{}{"pr": resp})
}

func (h *PRHandler) MergePR(w http.ResponseWriter, r *http.Request) {
	var req pr.MergeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.svc.Logger().Error(r.Context(), "Failed to decode MergeRequest", zap.Error(err))
		writeError(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	h.svc.Logger().Info(r.Context(), "MergePR request received", zap.String("pull_request_id", req.PullRequestID))

	resp, err := h.svc.MergePR(r.Context(), &req)
	if err != nil {
		h.svc.Logger().Error(r.Context(), "MergePR failed", zap.Error(err))
		switch err {
		case dto.ErrNotFound:
			writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		}
		return
	}

	h.svc.Logger().Info(r.Context(), "MergePR succeeded", zap.String("pull_request_id", resp.PullRequestID))
	writeJSON(w, http.StatusOK, map[string]interface{}{"pr": resp})
}

func (h *PRHandler) ReassignPR(w http.ResponseWriter, r *http.Request) {
	var req pr.ReassignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.svc.Logger().Error(r.Context(), "Failed to decode ReassignRequest", zap.Error(err))
		writeError(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	h.svc.Logger().Info(r.Context(), "ReassignPR request received",
		zap.String("pull_request_id", req.PullRequestID),
		zap.String("old_user_id", req.OldUserID),
	)

	resp, replacedBy, err := h.svc.ReassignReviewer(r.Context(), &req)
	if err != nil {
		h.svc.Logger().Error(r.Context(), "ReassignPR failed", zap.Error(err))
		switch err {
		case dto.ErrNotFound:
			writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		case dto.ErrPRMerged:
			writeError(w, http.StatusConflict, "PR_MERGED", err.Error())
		case dto.ErrNotAssigned:
			writeError(w, http.StatusConflict, "NOT_ASSIGNED", err.Error())
		case dto.ErrNoCandidate:
			writeError(w, http.StatusConflict, "NO_CANDIDATE", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		}
		return
	}

	h.svc.Logger().Info(r.Context(), "ReassignPR succeeded",
		zap.String("pull_request_id", resp.PullRequestID),
		zap.String("replaced_by", replacedBy),
	)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"pr":          resp,
		"replaced_by": replacedBy,
	})
}
