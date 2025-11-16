package postgres

import (
	"context"
	"database/sql"
	"errors"
	"pr_reviewer_assignment_service/internal/dto"
	"pr_reviewer_assignment_service/internal/entity"
	"pr_reviewer_assignment_service/pkg/logger"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type TeamRepository struct {
	db         *sqlx.DB
	sqlBuilder sq.StatementBuilderType
	logger     logger.Logger
}

func NewTeamRepository(db *sqlx.DB, logger logger.Logger) *TeamRepository {
	return &TeamRepository{
		db:         db,
		sqlBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		logger:     logger,
	}
}

func (r *TeamRepository) CreateTeam(ctx context.Context, team *entity.Team) error {
	r.logger.Info(ctx, "Creating team", zap.String("team_name", team.TeamName))

	query := r.sqlBuilder.Select("1").From("teams").Where("team_name = ?", team.TeamName).Limit(1)
	var exists int
	err := query.RunWith(r.db).QueryRowContext(ctx).Scan(&exists)
	if err == nil {
		r.logger.Warn(ctx, "Team already exists", zap.String("team_name", team.TeamName))
		return dto.ErrTeamExists
	} else if !errors.Is(err, sql.ErrNoRows) {
		r.logger.Error(ctx, "Failed to check existing team", zap.Error(err))
		return err
	}

	_, err = r.sqlBuilder.Insert("teams").Columns("team_name").Values(team.TeamName).RunWith(r.db).ExecContext(ctx)
	if err != nil {
		r.logger.Error(ctx, "Failed to insert team", zap.Error(err), zap.String("team_name", team.TeamName))
		return err
	}
	r.logger.Info(ctx, "Team inserted successfully", zap.String("team_name", team.TeamName))

	for _, member := range team.Members {
		_, err := r.sqlBuilder.
			Insert("users").
			Columns("user_id", "username", "team_name", "is_active").
			Values(member.UserID, member.Username, team.TeamName, member.IsActive).
			Suffix("ON CONFLICT (user_id) DO UPDATE SET username = EXCLUDED.username, is_active = EXCLUDED.is_active").
			RunWith(r.db).
			ExecContext(ctx)
		if err != nil {
			r.logger.Error(ctx, "Failed to insert/update user", zap.String("user_id", member.UserID), zap.Error(err))
			return err
		}
		r.logger.Info(ctx, "User inserted/updated", zap.String("user_id", member.UserID), zap.String("username", member.Username))
	}

	return nil
}

func (r *TeamRepository) GetTeamByName(ctx context.Context, name string) (*entity.Team, error) {
	name = strings.TrimSpace(name)
	r.logger.Info(ctx, "Fetching team by name", zap.String("team_name", name))

	var team entity.Team
	teamQuery := r.sqlBuilder.PlaceholderFormat(sq.Dollar).
		Select("team_name").
		From("teams").
		Where(sq.Eq{"team_name": name})

	teamSQL, teamArgs, err := teamQuery.ToSql()
	if err != nil {
		r.logger.Error(ctx, "Failed to build team query", zap.Error(err))
		return nil, err
	}

	err = r.db.QueryRowContext(ctx, teamSQL, teamArgs...).Scan(&team.TeamName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Warn(ctx, "Team not found", zap.String("team_name", name))
			return nil, dto.ErrNotFound
		}
		r.logger.Error(ctx, "Failed to fetch team", zap.Error(err))
		return nil, err
	}

	usersQuery := r.sqlBuilder.PlaceholderFormat(sq.Dollar).
		Select("user_id", "username", "team_name", "is_active").
		From("users").
		Where(sq.Eq{"team_name": name})

	usersSQL, usersArgs, err := usersQuery.ToSql()
	if err != nil {
		r.logger.Error(ctx, "Failed to build users query", zap.Error(err))
		return nil, err
	}

	rows, err := r.db.QueryxContext(ctx, usersSQL, usersArgs...)
	if err != nil {
		r.logger.Error(ctx, "Failed to fetch users", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var members []entity.User
	for rows.Next() {
		var u entity.User
		if err := rows.StructScan(&u); err != nil {
			r.logger.Error(ctx, "Failed to scan user row", zap.Error(err))
			return nil, err
		}
		members = append(members, u)
		r.logger.Debug(ctx, "User loaded", zap.String("user_id", u.UserID), zap.String("username", u.Username))
	}

	team.Members = members
	r.logger.Info(ctx, "Team loaded successfully", zap.String("team_name", team.TeamName), zap.Int("members_count", len(members)))

	return &team, nil
}
