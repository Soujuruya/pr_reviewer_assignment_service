package usecase

import (
	"context"
	"errors"
	"pr_reviewer_assignment_service/internal/dto"
	team "pr_reviewer_assignment_service/internal/dto/team"

	"pr_reviewer_assignment_service/internal/entity"
	"pr_reviewer_assignment_service/internal/repository/postgres"
	"pr_reviewer_assignment_service/pkg/logger"

	"go.uber.org/zap"
)

type TeamService struct {
	repo   *postgres.TeamRepository
	logger logger.Logger
}

func (s *TeamService) Logger() logger.Logger {
	return s.logger
}

func NewTeamService(repo *postgres.TeamRepository, logger logger.Logger) *TeamService {
	return &TeamService{repo: repo, logger: logger}
}

func (s *TeamService) CreateTeam(ctx context.Context, req *team.TeamRequest) (*team.TeamResponse, error) {
	s.logger.Info(ctx, "CreateTeam called", zap.String("team_name", req.TeamName))

	existing, err := s.repo.GetTeamByName(ctx, req.TeamName)
	if err != nil && !errors.Is(err, dto.ErrNotFound) {
		s.logger.Error(ctx, "Error checking existing team", zap.String("team_name", req.TeamName), zap.Error(err))
		return nil, err
	}
	if existing != nil {
		s.logger.Warn(ctx, "Team already exists", zap.String("team_name", req.TeamName))
		return nil, dto.ErrTeamExists
	}

	var members []entity.User
	for _, m := range req.Members {
		members = append(members, entity.User{
			UserID:   m.UserID,
			Username: m.Username,
			TeamName: req.TeamName,
			IsActive: m.IsActive,
		})
	}

	teamEntity := &entity.Team{
		TeamName: req.TeamName,
		Members:  members,
	}

	if err := s.repo.CreateTeam(ctx, teamEntity); err != nil {
		s.logger.Error(ctx, "Failed to create team", zap.String("team_name", req.TeamName), zap.Error(err))
		return nil, err
	}

	s.logger.Info(ctx, "Team created successfully", zap.String("team_name", req.TeamName), zap.Int("members_count", len(members)))

	resp := &team.TeamResponse{
		TeamName: teamEntity.TeamName,
		Members:  req.Members,
	}

	return resp, nil
}

func (s *TeamService) GetTeamByName(ctx context.Context, teamName string) (*team.TeamResponse, error) {
	s.logger.Info(ctx, "GetTeamByName called", zap.String("team_name", teamName))

	t, err := s.repo.GetTeamByName(ctx, teamName)
	if err != nil {
		s.logger.Error(ctx, "Team not found or error", zap.String("team_name", teamName), zap.Error(err))
		return nil, err
	}

	var members []team.TeamMember
	for _, u := range t.Members {
		members = append(members, team.TeamMember{
			UserID:   u.UserID,
			Username: u.Username,
			IsActive: u.IsActive,
		})
	}

	s.logger.Info(ctx, "Team retrieved successfully", zap.String("team_name", t.TeamName), zap.Int("members_count", len(members)))

	resp := &team.TeamResponse{
		TeamName: t.TeamName,
		Members:  members,
	}

	return resp, nil
}
