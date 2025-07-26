package pg

import (
	"auth/internal/domain/models"
	"auth/internal/repository"
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, email string, passHash []byte) (int64, error) {
	const op = "repository.user.postgres.Create"

	query := sq.Insert("users").
		Columns("email", "pass_hash").
		Values(email, passHash).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s: build query: %w", op, err)
	}

	var id int64
	if err := r.db.QueryRowContext(ctx, sqlStr, args...).Scan(&id); err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				return -1, fmt.Errorf("%s: %w", op, repository.ErrUserExists)
			}
		}

		return -1, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (r *UserRepository) Get(ctx context.Context, email string) (user models.User, err error) {
	const op = "repository.user.postgres.Get"

	query := sq.Select("id", "email", "pass_hash").
		From("users").
		Where(sq.Eq{"email": email}).
		PlaceholderFormat(sq.Dollar)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return user, fmt.Errorf("%s: build query: %w", op, err)
	}

	if err := r.db.QueryRowContext(ctx, sqlStr, args...).Scan(&user.ID, &user.Email, &user.PassHash); err != nil {
		if err == sql.ErrNoRows {
			return user, fmt.Errorf("%s: %w", op, repository.ErrUserNotFound)
		}
		return user, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}
