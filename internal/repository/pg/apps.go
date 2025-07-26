package pg

import (
	"auth/internal/domain/models"
	"auth/internal/repository"
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type AppRepository struct {
	db *sqlx.DB
}

func NewAppRepository(db *sqlx.DB) *AppRepository {
	return &AppRepository{db: db}
}

func (r *AppRepository) Get(ctx context.Context, appID int) (app models.App, err error) {
	const op = "repository.app.postgres.Get"

	query := sq.Select("*").
		From("apps").
		Where(sq.Eq{"id": appID}).
		PlaceholderFormat(sq.Dollar)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return app, fmt.Errorf("%s: build query: %w", op, err)
	}

	if err := r.db.QueryRowContext(ctx, sqlStr, args...).Scan(&app.ID, &app.Name, &app.AccessSecret, &app.RefreshSecret); err != nil {
		if err == sql.ErrNoRows {
			return app, fmt.Errorf("%s: %w", op, repository.ErrAppNotFound)
		}
		return app, fmt.Errorf("%s: %w", op, err)
	}
	return app, nil
}
