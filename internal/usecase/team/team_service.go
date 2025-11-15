package team

import (
	"context"
	"errors"
	"pr_reviewer_assignment_service/internal/dto"
	"pr_reviewer_assignment_service/internal/dto/team"
	"pr_reviewer_assignment_service/internal/entity"
	"pr_reviewer_assignment_service/internal/repository/postgres"
)

type TeamService struct {
	repo *postgres.TeamRepository
}

func NewTeamService(repo *postgres.TeamRepository) *TeamService {
	return &TeamService{repo: repo}
}

func (s *TeamService) CreateTeam(ctx context.Context, req *team.TeamRequest) (*team.TeamResponse, error) {
	existing, err := s.repo.GetTeamByName(ctx, req.TeamName)
	if err != nil && !errors.Is(err, dto.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
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
		return nil, err
	}

	resp := &team.TeamResponse{
		TeamName: teamEntity.TeamName,
		Members:  req.Members,
	}

	return resp, nil
}

func (s *TeamService) GetTeam(ctx context.Context, teamName string) (*team.TeamResponse, error) {
	t, err := s.repo.GetTeamByName(ctx, teamName)
	if err != nil {
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

	resp := &team.TeamResponse{
		TeamName: t.TeamName,
		Members:  members,
	}

	return resp, nil
}
