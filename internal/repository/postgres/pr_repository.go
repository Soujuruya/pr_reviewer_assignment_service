package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"pr_reviewer_assignment_service/internal/dto"
	"pr_reviewer_assignment_service/internal/entity"
	"pr_reviewer_assignment_service/pkg/logger"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type PRRepository struct {
	db     *sqlx.DB
	sb     sq.StatementBuilderType
	logger logger.Logger
}

func NewPRRepository(db *sqlx.DB, logger logger.Logger) *PRRepository {
	return &PRRepository{
		db:     db,
		sb:     sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		logger: logger,
	}
}

func (r *PRRepository) Create(ctx context.Context, pr *entity.PullRequest) error {
	r.logger.Info(ctx, "Creating Pull Request", zap.String("pr_id", pr.PullRequestID), zap.String("author_id", pr.AuthorID))
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		r.logger.Error(ctx, "Failed to begin transaction", zap.Error(err))
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	query := r.sb.Insert("pull_requests").
		Columns("pull_request_id", "pull_request_name", "author_id", "status", "created_at").
		Values(pr.PullRequestID, pr.Name, pr.AuthorID, pr.Status, pr.CreatedAt)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error(ctx, "Failed to build SQL", zap.Error(err))
		return err
	}

	if _, err = tx.ExecContext(ctx, sqlStr, args...); err != nil {
		r.logger.Error(ctx, "Failed to insert PR", zap.Error(err))
		return err
	}

	if len(pr.AssignedReviewers) > 0 {
		valueStrings := []string{}
		valueArgs := []interface{}{}
		i := 1
		for _, reviewer := range pr.AssignedReviewers {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", i, i+1))
			valueArgs = append(valueArgs, pr.PullRequestID, reviewer)
			i += 2
		}

		stmt := fmt.Sprintf(
			"INSERT INTO pull_request_reviewers (pull_request_id, user_id) VALUES %s",
			strings.Join(valueStrings, ","),
		)

		if _, err = tx.ExecContext(ctx, stmt, valueArgs...); err != nil {
			r.logger.Error(ctx, "Failed to insert reviewers", zap.Error(err))
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		r.logger.Error(ctx, "Failed to commit transaction", zap.Error(err))
		return err
	}

	r.logger.Info(ctx, "Pull Request created successfully", zap.String("pr_id", pr.PullRequestID))
	return nil
}

func (r *PRRepository) GetByID(ctx context.Context, prID string) (*entity.PullRequest, error) {
	r.logger.Debug(ctx, "GetByID called", zap.String("pr_id", prID))

	var pr entity.PullRequest
	err := r.db.GetContext(ctx, &pr, `
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE pull_request_id=$1
	`, prID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Warn(ctx, "PR not found", zap.String("pr_id", prID))
			return nil, dto.ErrNotFound
		}
		r.logger.Error(ctx, "Failed to get PR", zap.Error(err))
		return nil, err
	}

	var reviewers []string
	err = r.db.SelectContext(ctx, &reviewers, "SELECT user_id FROM pull_request_reviewers WHERE pull_request_id=$1", prID)
	if err != nil {
		r.logger.Error(ctx, "Failed to get reviewers", zap.String("pr_id", prID), zap.Error(err))
		return nil, err
	}
	pr.AssignedReviewers = reviewers

	r.logger.Debug(ctx, "GetByID successful", zap.String("pr_id", prID), zap.Int("reviewers_count", len(reviewers)))
	return &pr, nil
}

func (r *PRRepository) Merge(ctx context.Context, prID string, prEntity *entity.PullRequest) error {
	r.logger.Info(ctx, "Merging Pull Request", zap.String("pr_id", prID))

	query := r.sb.Update("pull_requests").
		Set("status", prEntity.Status).
		Set("merged_at", prEntity.MergedAt).
		Where(sq.Eq{"pull_request_id": prID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error(ctx, "Failed to build SQL for Merge", zap.Error(err))
		return err
	}

	res, err := r.db.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error(ctx, "Failed to execute Merge", zap.Error(err), zap.String("pr_id", prID))
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		r.logger.Warn(ctx, "No PR found to merge", zap.String("pr_id", prID))
		return dto.ErrNotFound
	}

	r.logger.Info(ctx, "Pull Request merged successfully", zap.String("pr_id", prID))
	return nil
}

func (r *PRRepository) ReassignReviewer(ctx context.Context, prID, oldUserID, newUserID string) (*entity.PullRequest, error) {
	r.logger.Info(ctx, "Reassigning reviewer", zap.String("pr_id", prID), zap.String("old_user_id", oldUserID), zap.String("new_user_id", newUserID))

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		r.logger.Error(ctx, "Failed to begin transaction", zap.Error(err))
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var exists bool
	err = tx.GetContext(ctx, &exists,
		"SELECT EXISTS(SELECT 1 FROM pull_request_reviewers WHERE pull_request_id=$1 AND user_id=$2)",
		prID, oldUserID,
	)
	if err != nil {
		r.logger.Error(ctx, "Failed to check if reviewer exists", zap.Error(err))
		return nil, err
	}
	if !exists {
		r.logger.Warn(ctx, "Old reviewer not assigned to PR", zap.String("pr_id", prID), zap.String("user_id", oldUserID))
		return nil, dto.ErrNotAssigned
	}

	_, err = tx.ExecContext(ctx,
		"DELETE FROM pull_request_reviewers WHERE pull_request_id=$1 AND user_id=$2",
		prID, oldUserID,
	)
	if err != nil {
		r.logger.Error(ctx, "Failed to remove old reviewer", zap.Error(err))
		return nil, err
	}

	if newUserID == "" {
		var candidates []string
		query := `
			SELECT user_id 
			FROM users 
			WHERE team_name = (
				SELECT team_name 
				FROM users 
				WHERE user_id = $1
			) 
			AND is_active = TRUE
			AND user_id != $1
		`
		err := tx.SelectContext(ctx, &candidates, query, oldUserID)
		if err != nil {
			r.logger.Error(ctx, "Failed to fetch candidate reviewers", zap.Error(err))
			return nil, err
		}

		if len(candidates) == 0 {
			r.logger.Warn(ctx, "No candidate reviewers available", zap.String("pr_id", prID))
			return nil, dto.ErrNoCandidate
		}

		newUserID = candidates[rand.Intn(len(candidates))]
		r.logger.Info(ctx, "Auto-selected new reviewer", zap.String("new_user_id", newUserID))
	}

	_, err = tx.ExecContext(ctx,
		"INSERT INTO pull_request_reviewers (pull_request_id, user_id) VALUES ($1, $2)",
		prID, newUserID,
	)
	if err != nil {
		r.logger.Error(ctx, "Failed to insert new reviewer", zap.Error(err))
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		r.logger.Error(ctx, "Failed to commit transaction for reassignment", zap.Error(err))
		return nil, err
	}

	prEntity := &entity.PullRequest{}
	err = r.db.GetContext(ctx, prEntity,
		`SELECT pull_request_id, pull_request_name, author_id, status
		 FROM pull_requests WHERE pull_request_id=$1`,
		prID,
	)
	if err != nil {
		r.logger.Error(ctx, "Failed to fetch PR after reassignment", zap.Error(err))
		return nil, err
	}

	var reviewers []string
	err = r.db.SelectContext(ctx, &reviewers,
		"SELECT user_id FROM pull_request_reviewers WHERE pull_request_id=$1",
		prID,
	)
	if err != nil {
		r.logger.Error(ctx, "Failed to fetch reviewers after reassignment", zap.Error(err))
		return nil, err
	}
	prEntity.AssignedReviewers = reviewers

	r.logger.Info(ctx, "Reviewer reassigned successfully", zap.String("pr_id", prID), zap.String("new_user_id", newUserID))
	return prEntity, nil
}

func (r *PRRepository) GetByReviewer(ctx context.Context, userID string) ([]*entity.PullRequest, error) {
	r.logger.Debug(ctx, "GetByReviewer called", zap.String("user_id", userID))

	query := `
		SELECT pr.pull_request_id,
		       pr.pull_request_name,
		       pr.author_id,
		       pr.status,
		       pr.created_at,
		       pr.merged_at
		FROM pull_requests pr
		JOIN pull_request_reviewers rev
		  ON pr.pull_request_id = rev.pull_request_id
		WHERE rev.user_id = $1
	`

	var prs []entity.PullRequest
	if err := r.db.SelectContext(ctx, &prs, query, userID); err != nil {
		r.logger.Error(ctx, "Failed to fetch PRs by reviewer", zap.Error(err))
		return nil, err
	}

	result := make([]*entity.PullRequest, 0, len(prs))
	for i := range prs {
		pr := &prs[i]
		var reviewers []string
		if err := r.db.SelectContext(ctx, &reviewers,
			"SELECT user_id FROM pull_request_reviewers WHERE pull_request_id=$1",
			pr.PullRequestID,
		); err != nil {
			r.logger.Error(ctx, "Failed to fetch reviewers for PR", zap.Error(err), zap.String("pr_id", pr.PullRequestID))
			return nil, err
		}
		pr.AssignedReviewers = reviewers
		result = append(result, pr)
	}

	r.logger.Debug(ctx, "GetByReviewer completed", zap.String("user_id", userID), zap.Int("prs_count", len(result)))
	return result, nil
}
