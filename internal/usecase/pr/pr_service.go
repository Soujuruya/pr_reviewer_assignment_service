package usecase

import (
	"context"
	"time"

	"pr_reviewer_assignment_service/internal/dto"
	"pr_reviewer_assignment_service/internal/dto/pr"
	"pr_reviewer_assignment_service/internal/entity"
	usecaseTeam "pr_reviewer_assignment_service/internal/usecase/team"
	usecaseUser "pr_reviewer_assignment_service/internal/usecase/user"
	"pr_reviewer_assignment_service/pkg/logger"

	"go.uber.org/zap"
)

type PRService struct {
	repo     PRRepository
	teamRepo usecaseTeam.TeamRepository
	userRepo usecaseUser.UserRepository
	logger   logger.Logger
}

func (s *PRService) Logger() logger.Logger {
	return s.logger
}

func NewPRService(repo PRRepository, teamRepo usecaseTeam.TeamRepository, userRepo usecaseUser.UserRepository, logger logger.Logger) *PRService {
	return &PRService{
		repo:     repo,
		teamRepo: teamRepo,
		userRepo: userRepo,
		logger:   logger,
	}
}

func (s *PRService) CreatePR(ctx context.Context, req *pr.CreatePRRequest) (*pr.PRResponse, error) {
	s.logger.Info(ctx, "CreatePR called",
		zap.String("pull_request_id", req.PullRequestID),
		zap.String("author_id", req.AuthorID),
	)

	existing, _ := s.repo.GetByID(ctx, req.PullRequestID)
	if existing != nil {
		s.logger.Warn(ctx, "PR already exists", zap.String("pull_request_id", req.PullRequestID))
		return nil, dto.ErrPRExists
	}

	author, err := s.userRepo.GetByID(ctx, req.AuthorID)
	if err != nil {
		s.logger.Error(ctx, "Author not found", zap.String("author_id", req.AuthorID), zap.Error(err))
		return nil, dto.ErrNotFound
	}

	team, err := s.teamRepo.GetTeamByName(ctx, author.TeamName)
	if err != nil {
		s.logger.Error(ctx, "Team not found", zap.String("team_name", author.TeamName), zap.Error(err))
		return nil, dto.ErrNotFound
	}

	candidates := []string{}
	for _, member := range team.Members {
		if member.IsActive && member.UserID != author.UserID {
			candidates = append(candidates, member.UserID)
		}
	}

	if len(candidates) > 2 {
		candidates = candidates[:2]
	}

	s.logger.Info(ctx, "Assigning reviewers", zap.Strings("reviewers", candidates))

	now := time.Now()
	prEntity := &entity.PullRequest{
		PullRequestID:     req.PullRequestID,
		Name:              req.PullRequestName,
		AuthorID:          req.AuthorID,
		Status:            entity.StatusOpen,
		AssignedReviewers: candidates,
		CreatedAt:         &now,
	}

	if err := s.repo.Create(ctx, prEntity); err != nil {
		s.logger.Error(ctx, "Failed to create PR", zap.Error(err))
		return nil, err
	}

	s.logger.Info(ctx, "PR created successfully", zap.String("pull_request_id", prEntity.PullRequestID))

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
	s.logger.Info(ctx, "MergePR called", zap.String("pull_request_id", req.PullRequestID))

	prEntity, err := s.repo.GetByID(ctx, req.PullRequestID)
	if err != nil {
		s.logger.Error(ctx, "PR not found", zap.String("pull_request_id", req.PullRequestID), zap.Error(err))
		return nil, err
	}

	if prEntity.Status == entity.StatusMerged {
		s.logger.Warn(ctx, "PR already merged", zap.String("pull_request_id", req.PullRequestID))
		return nil, dto.ErrPRMerged
	}

	now := time.Now()
	prEntity.Status = entity.StatusMerged
	prEntity.MergedAt = &now

	if err := s.repo.Merge(ctx, req.PullRequestID, prEntity); err != nil {
		s.logger.Error(ctx, "Failed to merge PR", zap.String("pull_request_id", req.PullRequestID), zap.Error(err))
		return nil, err
	}

	s.logger.Info(ctx, "PR merged successfully", zap.String("pull_request_id", prEntity.PullRequestID))

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
	s.logger.Info(ctx, "ReassignReviewer called",
		zap.String("pull_request_id", req.PullRequestID),
		zap.String("old_user_id", req.OldUserID),
	)

	prEntity, err := s.repo.ReassignReviewer(ctx, req.PullRequestID, req.OldUserID, "")
	if err != nil {
		s.logger.Error(ctx, "Failed to reassign reviewer",
			zap.String("pull_request_id", req.PullRequestID),
			zap.String("old_user_id", req.OldUserID),
			zap.Error(err),
		)
		return nil, "", err
	}

	var replacedBy string
	for _, r := range prEntity.AssignedReviewers {
		if r != req.OldUserID {
			replacedBy = r
			break
		}
	}

	s.logger.Info(ctx, "Reviewer reassigned",
		zap.String("pull_request_id", req.PullRequestID),
		zap.String("old_user_id", req.OldUserID),
		zap.String("new_user_id", replacedBy),
	)

	return &pr.PRResponse{
		PullRequestID:     prEntity.PullRequestID,
		PullRequestName:   prEntity.Name,
		AuthorID:          prEntity.AuthorID,
		Status:            string(prEntity.Status),
		AssignedReviewers: prEntity.AssignedReviewers,
	}, replacedBy, nil
}
