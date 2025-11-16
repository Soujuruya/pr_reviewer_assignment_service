package postgres

import (
	"context"
	"database/sql"
	"pr_reviewer_assignment_service/internal/dto"
	"pr_reviewer_assignment_service/internal/entity"
	"pr_reviewer_assignment_service/pkg/logger"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type UserRepository struct {
	db     *sqlx.DB
	sb     sq.StatementBuilderType
	logger logger.Logger
}

func NewUserRepository(db *sqlx.DB, logger logger.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		sb:     sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		logger: logger,
	}
}

func (r *UserRepository) GetByID(ctx context.Context, userID string) (*entity.User, error) {
	userID = strings.TrimSpace(userID)
	r.logger.Info(ctx, "Fetching user by ID", zap.String("user_id", userID))

	query := r.sb.Select("user_id", "username", "team_name", "is_active").
		From("users").
		Where(sq.Eq{"user_id": userID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error(ctx, "Failed to build GetByID query", zap.Error(err))
		return nil, err
	}

	var u entity.User
	if err := r.db.GetContext(ctx, &u, sqlStr, args...); err != nil {
		if err == sql.ErrNoRows {
			r.logger.Warn(ctx, "User not found", zap.String("user_id", userID))
			return nil, dto.ErrNotFound
		}
		r.logger.Error(ctx, "Failed to fetch user", zap.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "User fetched successfully", zap.String("user_id", u.UserID), zap.String("username", u.Username), zap.String("team_name", u.TeamName), zap.Bool("is_active", u.IsActive))
	return &u, nil
}

func (r *UserRepository) SetIsActive(ctx context.Context, userID string, isActive bool) (*entity.User, error) {
	r.logger.Info(ctx, "Updating user active status", zap.String("user_id", userID), zap.Bool("new_is_active", isActive))

	query := r.sb.Update("users").
		Set("is_active", isActive).
		Where(sq.Eq{"user_id": userID}).
		Suffix("RETURNING user_id, username, team_name, is_active")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error(ctx, "Failed to build SetIsActive query", zap.Error(err))
		return nil, err
	}

	var u entity.User
	if err := r.db.GetContext(ctx, &u, sqlStr, args...); err != nil {
		if err == sql.ErrNoRows {
			r.logger.Warn(ctx, "User not found when setting active", zap.String("user_id", userID))
			return nil, dto.ErrNotFound
		}
		r.logger.Error(ctx, "Failed to update user active status", zap.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "User active status updated", zap.String("user_id", u.UserID), zap.Bool("is_active", u.IsActive))
	return &u, nil
}
