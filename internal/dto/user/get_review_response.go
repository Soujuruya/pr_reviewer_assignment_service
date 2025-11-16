package user

import pr "pr_reviewer_assignment_service/internal/dto/pr"

type GetReviewResponse struct {
	UserID       string          `json:"user_id"`
	PullRequests []pr.PRShortDTO `json:"pull_requests"`
}
