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

type TeamRepository struct {
	db         *sqlx.DB
	sqlBuilder sq.StatementBuilderType
}

func NewTeamRepository(db *sqlx.DB) *TeamRepository {
	return &TeamRepository{
		db:         db,
		sqlBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *TeamRepository) CreateTeam(ctx context.Context, team *entity.Team) error {
	query := r.sqlBuilder.Select("1").From("teams").Where("team_name = ?", team.TeamName).Limit(1)
	var exists int
	err := query.RunWith(r.db).QueryRowContext(ctx).Scan(&exists)
	if err == nil {
		return dto.ErrTeamExists
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	_, err = r.sqlBuilder.Insert("teams").Columns("team_name").Values(team.TeamName).RunWith(r.db).ExecContext(ctx)
	if err != nil {
		return err
	}

	for _, member := range team.Members {
		_, err := r.sqlBuilder.
			Insert("users").
			Columns("user_id", "username", "team_name", "is_active").
			Values(member.UserID, member.Username, team.TeamName, member.IsActive).
			Suffix("ON CONFLICT (user_id) DO UPDATE SET username = EXCLUDED.username, is_active = EXCLUDED.is_active").
			RunWith(r.db).
			ExecContext(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *TeamRepository) GetTeamByName(ctx context.Context, name string) (*entity.Team, error) {
	var team entity.Team

	query := r.sqlBuilder.Select("team_name").From("teams").Where("team_name = ?", name)
	err := query.RunWith(r.db).QueryRowContext(ctx).Scan(&team.TeamName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, dto.ErrNotFound
		}
		return nil, err
	}

	var members []entity.User
	usersQuery := r.sqlBuilder.
		Select("user_id", "username", "team_name", "is_active").
		From("users").
		Where("team_name = ?", name)

	sqlStr, args, err := usersQuery.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryxContext(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var u entity.User
		if err := rows.StructScan(&u); err != nil {
			return nil, err
		}
		members = append(members, u)
	}

	team.Members = members
	return &team, nil
}
