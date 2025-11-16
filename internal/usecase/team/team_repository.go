package usecase

import (
	"context"
	"pr_reviewer_assignment_service/internal/entity"
)

type TeamRepository interface {
	CreateTeam(ctx context.Context, team *entity.Team) error
	GetTeamByName(ctx context.Context, teamName string) (*entity.Team, error)
}
