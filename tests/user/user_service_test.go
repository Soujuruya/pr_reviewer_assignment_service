package user_test

import (
	"context"
	"errors"
	"testing"

	"pr_reviewer_assignment_service/internal/dto/user"
	"pr_reviewer_assignment_service/internal/entity"
	usecase "pr_reviewer_assignment_service/internal/usecase/user"
	mockLogger "pr_reviewer_assignment_service/mocks/logger"
	mockUser "pr_reviewer_assignment_service/mocks/user"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestUserService_SetActive(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockUser.NewMockUserRepository(ctrl)
	mockPRRepo := mockUser.NewMockPRGetter(ctrl)
	mockLogger := mockLogger.NewMockLogger()

	svc := usecase.NewUserService(mockRepo, mockPRRepo, mockLogger)

	t.Run("success", func(t *testing.T) {
		req := &user.SetIsActiveRequest{
			UserID:   "uuid-123",
			IsActive: true,
		}

		mockRepo.EXPECT().
			SetIsActive(ctx, req.UserID, req.IsActive).
			Return(&entity.User{
				UserID:   req.UserID,
				Username: "testuser",
				TeamName: "team1",
				IsActive: true,
			}, nil)

		resp, err := svc.SetActive(ctx, req)
		require.NoError(t, err)
		require.Equal(t, req.UserID, resp.UserID)
		require.Equal(t, true, resp.IsActive)
	})

	t.Run("repo error", func(t *testing.T) {
		req := &user.SetIsActiveRequest{
			UserID:   "uuid-456",
			IsActive: false,
		}

		mockRepo.EXPECT().
			SetIsActive(ctx, req.UserID, req.IsActive).
			Return(nil, errors.New("db error"))

		resp, err := svc.SetActive(ctx, req)
		require.Nil(t, resp)
		require.Error(t, err)
	})
}

func TestUserService_GetReview(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockUser.NewMockUserRepository(ctrl)
	mockPRRepo := mockUser.NewMockPRGetter(ctrl)
	mockLogger := mockLogger.NewMockLogger()

	svc := usecase.NewUserService(mockRepo, mockPRRepo, mockLogger)

	t.Run("success", func(t *testing.T) {
		userID := "uuid-123"
		mockPRRepo.EXPECT().
			GetByReviewer(ctx, userID).
			Return([]*entity.PullRequest{
				{
					PullRequestID: "pr-1",
					Name:          "Fix bug",
					AuthorID:      "author-1",
					Status:        entity.StatusOpen,
				},
				{
					PullRequestID: "pr-2",
					Name:          "Add feature",
					AuthorID:      "author-2",
					Status:        entity.StatusMerged,
				},
			}, nil)

		resp, err := svc.GetReview(ctx, userID)
		require.NoError(t, err)
		require.Equal(t, userID, resp.UserID)
		require.Len(t, resp.PullRequests, 2)
		require.Equal(t, "pr-1", resp.PullRequests[0].PullRequestID)
		require.Equal(t, "pr-2", resp.PullRequests[1].PullRequestID)
	})

	t.Run("repo error", func(t *testing.T) {
		userID := "uuid-456"
		mockPRRepo.EXPECT().
			GetByReviewer(ctx, userID).
			Return(nil, errors.New("db error"))

		resp, err := svc.GetReview(ctx, userID)
		require.Nil(t, resp)
		require.Error(t, err)
	})
}
