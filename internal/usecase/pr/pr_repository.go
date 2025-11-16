package usecase

import (
	"context"
	"pr_reviewer_assignment_service/internal/entity"
)

// internal/usecase/pr/pr_service.go
type PRRepository interface {
	GetByID(ctx context.Context, prID string) (*entity.PullRequest, error)
	Create(ctx context.Context, pr *entity.PullRequest) error
	Merge(ctx context.Context, prID string, pr *entity.PullRequest) error
	ReassignReviewer(ctx context.Context, prID, oldUserID, newUserID string) (*entity.PullRequest, error)
	GetByReviewer(ctx context.Context, userID string) ([]*entity.PullRequest, error)
}
