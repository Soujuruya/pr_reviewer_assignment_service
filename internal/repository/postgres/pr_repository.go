package postgres

import (
	"context"
	"database/sql"
	"errors"

	"pr_reviewer_assignment_service/internal/dto"
	"pr_reviewer_assignment_service/internal/entity"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type PRRepository struct {
	db *sqlx.DB
	sb sq.StatementBuilderType
}

// Конструктор
func NewPRRepository(db *sqlx.DB) *PRRepository {
	return &PRRepository{
		db: db,
		sb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

// Create создаёт новый PR
func (r *PRRepository) Create(ctx context.Context, pr *entity.PullRequest) error {
	query := r.sb.Insert("pull_requests").
		Columns("pull_request_id", "pull_request_name", "author_id", "status", "created_at").
		Values(pr.PullRequestID, pr.Name, pr.AuthorID, pr.Status, pr.CreatedAt)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		return err
	}

	// Назначение ревьюверов
	for _, reviewer := range pr.AssignedReviewers {
		_, err := r.db.ExecContext(ctx,
			"INSERT INTO pull_request_reviewers (pull_request_id, user_id) VALUES ($1, $2)",
			pr.PullRequestID, reviewer,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetByID возвращает PR по ID
func (r *PRRepository) GetByID(ctx context.Context, prID string) (*entity.PullRequest, error) {
	var pr entity.PullRequest

	// Получаем основной PR
	err := r.db.GetContext(ctx, &pr, `
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE pull_request_id=$1
	`, prID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, dto.ErrNotFound
		}
		return nil, err
	}

	// Получаем назначенных ревьюверов
	var reviewers []string
	err = r.db.SelectContext(ctx, &reviewers,
		"SELECT user_id FROM pull_request_reviewers WHERE pull_request_id=$1", prID)
	if err != nil {
		return nil, err
	}

	pr.AssignedReviewers = reviewers
	return &pr, nil
}

// Merge помечает PR как MERGED
func (r *PRRepository) Merge(ctx context.Context, prID string, prEntity *entity.PullRequest) error {
	query := r.sb.Update("pull_requests").
		Set("status", prEntity.Status).
		Set("merged_at", prEntity.MergedAt).
		Where(sq.Eq{"pull_request_id": prID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return err
	}

	res, err := r.db.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return dto.ErrNotFound
	}

	return nil
}

// ReassignReviewer заменяет старого ревьювера на нового
func (r *PRRepository) ReassignReviewer(ctx context.Context, prID, oldUserID, newUserID string) (*entity.PullRequest, error) {
	// Проверяем, что старый ревьювер назначен
	var exists bool
	err := r.db.GetContext(ctx, &exists,
		"SELECT EXISTS(SELECT 1 FROM pull_request_reviewers WHERE pull_request_id=$1 AND user_id=$2)",
		prID, oldUserID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, dto.ErrNotAssigned
	}

	// Обновляем на нового
	_, err = r.db.ExecContext(ctx,
		"UPDATE pull_request_reviewers SET user_id=$1 WHERE pull_request_id=$2 AND user_id=$3",
		newUserID, prID, oldUserID)
	if err != nil {
		return nil, err
	}

	// Возвращаем обновлённый PR
	return r.GetByID(ctx, prID)
}

// GetByReviewer возвращает список PR, где пользователь назначен ревьювером
func (r *PRRepository) GetByReviewer(ctx context.Context, userID string) ([]*entity.PullRequest, error) {
	var prIDs []string
	err := r.db.SelectContext(ctx, &prIDs,
		"SELECT pull_request_id FROM pull_request_reviewers WHERE user_id=$1", userID)
	if err != nil {
		return nil, err
	}

	prs := make([]*entity.PullRequest, 0, len(prIDs))
	for _, prID := range prIDs {
		pr, err := r.GetByID(ctx, prID)
		if err != nil {
			continue
		}
		prs = append(prs, pr)
	}

	return prs, nil
}
