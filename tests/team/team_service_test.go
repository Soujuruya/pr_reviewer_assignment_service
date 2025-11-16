package team_test

import (
	"context"
	"testing"

	"pr_reviewer_assignment_service/internal/dto"
	teamDTO "pr_reviewer_assignment_service/internal/dto/team"
	entity "pr_reviewer_assignment_service/internal/entity"
	usecase "pr_reviewer_assignment_service/internal/usecase/team"
	mockLogger "pr_reviewer_assignment_service/mocks/logger"
	mockTeam "pr_reviewer_assignment_service/mocks/team"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestTeamService_CreateTeam_AlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	repo := mockTeam.NewMockTeamRepository(ctrl)
	logger := mockLogger.NewMockLogger()

	service := usecase.NewTeamService(repo, logger)

	req := &teamDTO.TeamRequest{
		TeamName: "team-1",
		Members: []teamDTO.TeamMember{
			{UserID: "uuid-1", Username: "user1", IsActive: true},
		},
	}

	repo.EXPECT().GetTeamByName(ctx, req.TeamName).
		Return(&entity.Team{TeamName: req.TeamName}, nil)

	resp, err := service.CreateTeam(ctx, req)
	require.Nil(t, resp)
	require.ErrorIs(t, err, dto.ErrTeamExists)
}

func TestTeamService_CreateTeam_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	repo := mockTeam.NewMockTeamRepository(ctrl)
	logger := mockLogger.NewMockLogger()

	service := usecase.NewTeamService(repo, logger)

	req := &teamDTO.TeamRequest{
		TeamName: "team-1",
		Members: []teamDTO.TeamMember{
			{UserID: "uuid-1", Username: "user1", IsActive: true},
			{UserID: "uuid-2", Username: "user2", IsActive: false},
		},
	}

	repo.EXPECT().GetTeamByName(ctx, req.TeamName).
		Return(nil, dto.ErrNotFound)

	repo.EXPECT().CreateTeam(ctx, gomock.Any()).Return(nil)

	resp, err := service.CreateTeam(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, req.TeamName, resp.TeamName)
	require.Len(t, resp.Members, 2)
}

func TestTeamService_GetTeamByName_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	repo := mockTeam.NewMockTeamRepository(ctrl)
	logger := mockLogger.NewMockLogger()

	service := usecase.NewTeamService(repo, logger)

	teamName := "team-1"

	entityTeam := &entity.Team{
		TeamName: teamName,
		Members: []entity.User{
			{UserID: "uuid-1", Username: "user1", IsActive: true},
			{UserID: "uuid-2", Username: "user2", IsActive: false},
		},
	}

	repo.EXPECT().GetTeamByName(ctx, teamName).Return(entityTeam, nil)

	resp, err := service.GetTeamByName(ctx, teamName)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, teamName, resp.TeamName)
	require.Len(t, resp.Members, 2)
}

func TestTeamService_GetTeamByName_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	repo := mockTeam.NewMockTeamRepository(ctrl)
	logger := mockLogger.NewMockLogger()

	service := usecase.NewTeamService(repo, logger)

	teamName := "team-1"

	repo.EXPECT().GetTeamByName(ctx, teamName).Return(nil, dto.ErrNotFound)

	resp, err := service.GetTeamByName(ctx, teamName)
	require.Nil(t, resp)
	require.ErrorIs(t, err, dto.ErrNotFound)
}
