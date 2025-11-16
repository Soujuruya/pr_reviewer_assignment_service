package usecase

import (
	"context"
	"pr_reviewer_assignment_service/internal/entity"
)

type UserRepository interface {
	GetByID(ctx context.Context, userID string) (*entity.User, error)
	SetIsActive(ctx context.Context, userID string, isActive bool) (*entity.User, error)
}

type PRGetter interface {
	GetByReviewer(ctx context.Context, userID string) ([]*entity.PullRequest, error)
}
