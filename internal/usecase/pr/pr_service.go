package pr

import (
	"context"
	"time"

	"pr_reviewer_assignment_service/internal/dto"
	"pr_reviewer_assignment_service/internal/dto/pr"
	"pr_reviewer_assignment_service/internal/entity"
)

type PRService struct {
	repo PRRepository
}

func NewPRService(repo PRRepository) *PRService {
	return &PRService{repo: repo}
}

func (s *PRService) CreatePR(ctx context.Context, req *pr.CreatePRRequest) (*pr.PRResponse, error) {
	existing, _ := s.repo.GetByID(ctx, req.PullRequestID)
	if existing != nil {
		return nil, dto.ErrPRExists
	}

	now := time.Now()
	prEntity := &entity.PullRequest{
		PullRequestID:     req.PullRequestID,
		Name:              req.PullRequestName,
		AuthorID:          req.AuthorID,
		Status:            entity.StatusOpen,
		AssignedReviewers: []string{},
		CreatedAt:         &now,
	}

	if err := s.repo.Create(ctx, prEntity); err != nil {
		return nil, err
	}

	return &pr.PRResponse{
		PullRequestID:     prEntity.PullRequestID,
		PullRequestName:   prEntity.Name,
		AuthorID:          prEntity.AuthorID,
		Status:            string(prEntity.Status),
		AssignedReviewers: prEntity.AssignedReviewers,
		CreatedAt:         nil,
	}, nil
}

func (s *PRService) MergePR(ctx context.Context, req *pr.MergeRequest) (*pr.PRResponse, error) {
	prEntity, err := s.repo.GetByID(ctx, req.PullRequestID)
	if err != nil {
		return nil, err
	}

	if prEntity.Status == entity.StatusMerged {
		return nil, dto.ErrPRMerged
	}

	now := time.Now()
	prEntity.Status = entity.StatusMerged
	prEntity.MergedAt = &now

	if err := s.repo.Merge(ctx, req.PullRequestID, prEntity); err != nil {
		return nil, err
	}

	return &pr.PRResponse{
		PullRequestID:     prEntity.PullRequestID,
		PullRequestName:   prEntity.Name,
		AuthorID:          prEntity.AuthorID,
		Status:            string(prEntity.Status),
		AssignedReviewers: prEntity.AssignedReviewers,
		MergedAt:          nil,
	}, nil
}

func (s *PRService) ReassignReviewer(ctx context.Context, req *pr.ReassignRequest) (*pr.PRResponse, string, error) {
	prEntity, err := s.repo.ReassignReviewer(ctx, req.PullRequestID, req.OldUserID, "")
	if err != nil {
		return nil, "", err
	}

	return &pr.PRResponse{
		PullRequestID:     prEntity.PullRequestID,
		PullRequestName:   prEntity.Name,
		AuthorID:          prEntity.AuthorID,
		Status:            string(prEntity.Status),
		AssignedReviewers: prEntity.AssignedReviewers,
	}, "", nil
}
