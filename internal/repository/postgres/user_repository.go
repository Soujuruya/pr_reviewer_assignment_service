package postgres

import (
	"context"
	"database/sql"
	"pr_reviewer_assignment_service/internal/dto"
	"pr_reviewer_assignment_service/internal/entity"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
	sb sq.StatementBuilderType
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{
		db: db,
		sb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *UserRepository) GetByID(ctx context.Context, userID string) (*entity.User, error) {
	query := r.sb.Select("user_id", "username", "team_name", "is_active").
		From("users").
		Where(sq.Eq{"user_id": userID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var u entity.User
	if err := r.db.GetContext(ctx, &u, sqlStr, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, dto.ErrNotFound
		}
		return nil, err
	}

	return &u, nil
}

func (r *UserRepository) SetIsActive(ctx context.Context, userID string, isActive bool) (*entity.User, error) {
	query := r.sb.Update("users").
		Set("is_active", isActive).
		Where(sq.Eq{"user_id": userID}).
		Suffix("RETURNING user_id, username, team_name, is_active")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var u entity.User
	if err := r.db.GetContext(ctx, &u, sqlStr, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, dto.ErrNotFound
		}
		return nil, err
	}

	return &u, nil
}
