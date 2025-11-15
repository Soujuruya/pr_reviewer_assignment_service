package user

import (
	"context"
	"pr_reviewer_assignment_service/internal/dto/pr"
	"pr_reviewer_assignment_service/internal/dto/user"
	"pr_reviewer_assignment_service/internal/repository/postgres"
)

type UserService struct {
	repo   *postgres.UserRepository
	prRepo PRGetter
}

func NewUserService(repo *postgres.UserRepository, prRepo PRGetter) *UserService {
	return &UserService{repo: repo, prRepo: prRepo}
}

func (s *UserService) SetActive(ctx context.Context, req *user.SetIsActiveRequest) (*user.UserResponse, error) {
	u, err := s.repo.SetIsActive(ctx, req.UserID, req.IsActive)
	if err != nil {
		return nil, err
	}

	return &user.UserResponse{
		UserID:   u.UserID,
		Username: u.Username,
		TeamName: u.TeamName,
		IsActive: u.IsActive,
	}, nil
}

func (s *UserService) GetReview(ctx context.Context, userID string) (*user.GetReviewResponse, error) {
	prs, err := s.prRepo.GetPRsByReviewer(ctx, userID)
	if err != nil {
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

	return &user.GetReviewResponse{
		UserID:       userID,
		PullRequests: shortList,
	}, nil
}
