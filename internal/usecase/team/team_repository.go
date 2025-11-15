package team

import (
	"context"
	"pr_reviewer_assignment_service/internal/entity"
)

type Repository interface {
	Create(ctx context.Context, team *entity.Team) error
	GetByName(ctx context.Context, teamName string) (*entity.Team, error)
}
