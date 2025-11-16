package usecase

import (
	"context"

	"pr_reviewer_assignment_service/internal/dto/pr"
	"pr_reviewer_assignment_service/internal/dto/user"
	"pr_reviewer_assignment_service/internal/repository/postgres"
	"pr_reviewer_assignment_service/pkg/logger"

	"go.uber.org/zap"
)

type UserService struct {
	repo   *postgres.UserRepository
	prRepo PRGetter
	logger logger.Logger
}

func (s *UserService) Logger() logger.Logger {
	return s.logger
}

func NewUserService(repo *postgres.UserRepository, prRepo PRGetter, logger logger.Logger) *UserService {
	return &UserService{repo: repo, prRepo: prRepo, logger: logger}
}

func (s *UserService) SetActive(ctx context.Context, req *user.SetIsActiveRequest) (*user.UserResponse, error) {
	s.logger.Info(ctx, "SetActive called", zap.String("user_id", req.UserID), zap.Bool("is_active", req.IsActive))

	u, err := s.repo.SetIsActive(ctx, req.UserID, req.IsActive)
	if err != nil {
		s.logger.Error(ctx, "Failed to set active status", zap.String("user_id", req.UserID), zap.Error(err))
		return nil, err
	}

	s.logger.Info(ctx, "SetActive successful", zap.String("user_id", u.UserID), zap.Bool("is_active", u.IsActive))

	return &user.UserResponse{
		UserID:   u.UserID,
		Username: u.Username,
		TeamName: u.TeamName,
		IsActive: u.IsActive,
	}, nil
}

func (s *UserService) GetReview(ctx context.Context, userID string) (*user.GetReviewResponse, error) {
	s.logger.Info(ctx, "GetReview called", zap.String("user_id", userID))

	prs, err := s.prRepo.GetByReviewer(ctx, userID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get reviews for user", zap.String("user_id", userID), zap.Error(err))
		return nil, err
	}

	shortList := make([]pr.PRShortDTO, 0, len(prs))
	for _, p := range prs {
		shortList = append(shortList, pr.PRShortDTO{
			PullRequestID:   p.PullRequestID,
			PullRequestName: p.Name,
			AuthorID:        p.AuthorID,
			Status:          string(p.Status),
		})
	}

	s.logger.Info(ctx, "GetReview successful", zap.String("user_id", userID), zap.Int("pull_requests_count", len(shortList)))

	return &user.GetReviewResponse{
		UserID:       userID,
		PullRequests: shortList,
	}, nil
}
