package pr_test

import (
	"context"
	"errors"
	"testing"

	"pr_reviewer_assignment_service/internal/dto"
	dtoPR "pr_reviewer_assignment_service/internal/dto/pr"
	entity "pr_reviewer_assignment_service/internal/entity"
	usecasePr "pr_reviewer_assignment_service/internal/usecase/pr"
	mockLogger "pr_reviewer_assignment_service/mocks/logger"
	mockPR "pr_reviewer_assignment_service/mocks/pr"
	mockTeam "pr_reviewer_assignment_service/mocks/team"
	mockUser "pr_reviewer_assignment_service/mocks/user"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestCreatePR_AlreadyExists(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	repo := mockPR.NewMockPRRepository(ctrl)
	teamRepo := mockTeam.NewMockTeamRepository(ctrl)
	userRepo := mockUser.NewMockUserRepository(ctrl)
	logger := mockLogger.NewMockLogger()

	svc := usecasePr.NewPRService(repo, teamRepo, userRepo, logger)

	req := &dtoPR.CreatePRRequest{
		PullRequestID:   "f0375e25-ffba-4c6f-885d-6c3b8350d81f",
		PullRequestName: "Test",
		AuthorID:        "40ef164f-5bd3-4196-b1e3-c9ed9a652579",
	}

	repo.EXPECT().GetByID(ctx, req.PullRequestID).
		Return(&entity.PullRequest{}, nil)

	resp, err := svc.CreatePR(ctx, req)

	require.Nil(t, resp)
	require.ErrorIs(t, err, dto.ErrPRExists)
}

func TestCreatePR_AuthorNotFound(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	repo := mockPR.NewMockPRRepository(ctrl)
	teamRepo := mockTeam.NewMockTeamRepository(ctrl)
	userRepo := mockUser.NewMockUserRepository(ctrl)
	logger := mockLogger.NewMockLogger()

	svc := usecasePr.NewPRService(repo, teamRepo, userRepo, logger)

	req := &dtoPR.CreatePRRequest{
		PullRequestID: "f0375e25-ffba-4c6f-885d-6c3b8350d81f",
		AuthorID:      "40ef164f-5bd3-4196-b1e3-c9ed9a652579",
	}

	repo.EXPECT().GetByID(ctx, req.PullRequestID).Return(nil, nil)

	userRepo.EXPECT().GetByID(ctx, req.AuthorID).Return(nil, errors.New("not found"))

	resp, err := svc.CreatePR(ctx, req)

	require.Nil(t, resp)
	require.ErrorIs(t, err, dto.ErrNotFound)
}

func TestCreatePR_TeamNotFound(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	repo := mockPR.NewMockPRRepository(ctrl)
	teamRepo := mockTeam.NewMockTeamRepository(ctrl)
	userRepo := mockUser.NewMockUserRepository(ctrl)
	logger := mockLogger.NewMockLogger()

	svc := usecasePr.NewPRService(repo, teamRepo, userRepo, logger)

	req := &dtoPR.CreatePRRequest{
		PullRequestID: "f0375e25-ffba-4c6f-885d-6c3b8350d81f",
		AuthorID:      "40ef164f-5bd3-4196-b1e3-c9ed9a652579",
	}

	user := &entity.User{
		UserID:   req.AuthorID,
		TeamName: "Backend",
	}

	repo.EXPECT().GetByID(ctx, req.PullRequestID).Return(nil, nil)

	userRepo.EXPECT().GetByID(ctx, req.AuthorID).Return(user, nil)

	teamRepo.EXPECT().
		GetTeamByName(ctx, "Backend").
		Return(nil, errors.New("not found"))

	resp, err := svc.CreatePR(ctx, req)

	require.Nil(t, resp)
	require.ErrorIs(t, err, dto.ErrNotFound)
}

func TestCreatePR_RepoCreateError(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	repo := mockPR.NewMockPRRepository(ctrl)
	teamRepo := mockTeam.NewMockTeamRepository(ctrl)
	userRepo := mockUser.NewMockUserRepository(ctrl)
	logger := mockLogger.NewMockLogger()

	svc := usecasePr.NewPRService(repo, teamRepo, userRepo, logger)

	req := &dtoPR.CreatePRRequest{
		PullRequestID: "f0375e25-ffba-4c6f-885d-6c3b8350d81f",
		AuthorID:      "40ef164f-5bd3-4196-b1e3-c9ed9a652579",
	}

	user := &entity.User{UserID: req.AuthorID, TeamName: "Backend"}

	team := &entity.Team{
		TeamName: "Backend",
		Members:  []entity.User{},
	}

	repo.EXPECT().GetByID(ctx, req.PullRequestID).Return(nil, nil)
	userRepo.EXPECT().GetByID(ctx, req.AuthorID).Return(user, nil)
	teamRepo.EXPECT().GetTeamByName(ctx, "Backend").Return(team, nil)

	repo.EXPECT().Create(ctx, gomock.Any()).Return(errors.New("db fail"))

	resp, err := svc.CreatePR(ctx, req)

	require.Nil(t, resp)
	require.EqualError(t, err, "db fail")
}

func TestMergePR_NotFound(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	repo := mockPR.NewMockPRRepository(ctrl)
	logger := mockLogger.NewMockLogger()

	svc := usecasePr.NewPRService(repo, nil, nil, logger)

	req := &dtoPR.MergeRequest{PullRequestID: "f0375e25-ffba-4c6f-885d-6c3b8350d81f"}

	repo.EXPECT().
		GetByID(ctx, req.PullRequestID).
		Return(nil, errors.New("not found"))

	resp, err := svc.MergePR(ctx, req)

	require.Nil(t, resp)
	require.EqualError(t, err, "not found")
}

func TestMergePR_AlreadyMerged(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	repo := mockPR.NewMockPRRepository(ctrl)
	logger := mockLogger.NewMockLogger()

	svc := usecasePr.NewPRService(repo, nil, nil, logger)

	prEntity := &entity.PullRequest{
		PullRequestID: "f0375e25-ffba-4c6f-885d-6c3b8350d81f",
		Status:        entity.StatusMerged,
	}

	req := &dtoPR.MergeRequest{PullRequestID: prEntity.PullRequestID}

	repo.EXPECT().GetByID(ctx, prEntity.PullRequestID).Return(prEntity, nil)

	resp, err := svc.MergePR(ctx, req)

	require.Nil(t, resp)
	require.ErrorIs(t, err, dto.ErrPRMerged)
}

func TestMergePR_RepoError(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	repo := mockPR.NewMockPRRepository(ctrl)
	logger := mockLogger.NewMockLogger()

	svc := usecasePr.NewPRService(repo, nil, nil, logger)

	prEntity := &entity.PullRequest{
		PullRequestID: "f0375e25-ffba-4c6f-885d-6c3b8350d81f",
		Status:        entity.StatusOpen,
	}

	req := &dtoPR.MergeRequest{PullRequestID: "f0375e25-ffba-4c6f-885d-6c3b8350d81f"}

	repo.EXPECT().GetByID(ctx, "f0375e25-ffba-4c6f-885d-6c3b8350d81f").Return(prEntity, nil)
	repo.EXPECT().Merge(ctx, "f0375e25-ffba-4c6f-885d-6c3b8350d81f", gomock.Any()).
		Return(errors.New("merge failed"))

	resp, err := svc.MergePR(ctx, req)

	require.Nil(t, resp)
	require.EqualError(t, err, "merge failed")
}

func TestMergePR_Success(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	repo := mockPR.NewMockPRRepository(ctrl)
	logger := mockLogger.NewMockLogger()

	svc := usecasePr.NewPRService(repo, nil, nil, logger)

	prEntity := &entity.PullRequest{
		PullRequestID: "f0375e25-ffba-4c6f-885d-6c3b8350d81f",
		Status:        entity.StatusOpen,
	}

	req := &dtoPR.MergeRequest{PullRequestID: "f0375e25-ffba-4c6f-885d-6c3b8350d81f"}

	repo.EXPECT().GetByID(ctx, "f0375e25-ffba-4c6f-885d-6c3b8350d81f").Return(prEntity, nil)
	repo.EXPECT().Merge(ctx, "f0375e25-ffba-4c6f-885d-6c3b8350d81f", gomock.Any()).Return(nil)

	resp, err := svc.MergePR(ctx, req)

	require.NoError(t, err)
	require.Equal(t, "f0375e25-ffba-4c6f-885d-6c3b8350d81f", resp.PullRequestID)
	require.Equal(t, string(entity.StatusMerged), resp.Status)
}

func TestReassignReviewer_RepoError(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	repo := mockPR.NewMockPRRepository(ctrl)
	logger := mockLogger.NewMockLogger()

	svc := usecasePr.NewPRService(repo, nil, nil, logger)

	req := &dtoPR.ReassignRequest{
		PullRequestID: "f0375e25-ffba-4c6f-885d-6c3b8350d81f",
		OldUserID:     "old",
	}

	repo.EXPECT().
		ReassignReviewer(ctx, req.PullRequestID, "old", "").
		Return(nil, errors.New("fail"))

	resp, replacedBy, err := svc.ReassignReviewer(ctx, req)

	require.Nil(t, resp)
	require.Empty(t, replacedBy)
	require.EqualError(t, err, "fail")
}

func TestReassignReviewer_Success(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	repo := mockPR.NewMockPRRepository(ctrl)
	logger := mockLogger.NewMockLogger()

	svc := usecasePr.NewPRService(repo, nil, nil, logger)

	prEntity := &entity.PullRequest{
		PullRequestID: "f0375e25-ffba-4c6f-885d-6c3b8350d81f",
		Name:          "PR",
		AuthorID:      "40ef164f-5bd3-4196-b1e3-c9ed9a652579",
		Status:        entity.StatusOpen,
		AssignedReviewers: []string{
			"new",
			"old",
		},
	}

	req := &dtoPR.ReassignRequest{
		PullRequestID: "f0375e25-ffba-4c6f-885d-6c3b8350d81f",
		OldUserID:     "old",
	}

	repo.EXPECT().
		ReassignReviewer(ctx, "f0375e25-ffba-4c6f-885d-6c3b8350d81f", "old", "").
		Return(prEntity, nil)

	resp, replacedBy, err := svc.ReassignReviewer(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "new", replacedBy)
}
