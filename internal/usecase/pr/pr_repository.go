package pr

import (
	"context"
	"pr_reviewer_assignment_service/internal/entity"
)

type PRRepository interface {
	Create(ctx context.Context, pr *entity.PullRequest) error
	GetByID(ctx context.Context, prID string) (*entity.PullRequest, error)
	Merge(ctx context.Context, prID string, mergedAt *entity.PullRequest) error
	ReassignReviewer(ctx context.Context, prID string, oldUserID, newUserID string) (*entity.PullRequest, error)
	GetByReviewer(ctx context.Context, userID string) ([]*entity.PullRequest, error)
}
