package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Config struct {
	UserName string `env:"POSTGRES_USER" env-default:"root"`
	Password string `env:"POSTGRES_PASSWORD" env-default:"111"`
	Host     string `env:"POSTGRES_HOST" env-default:"localhost"`
	Port     int    `env:"POSTGRES_PORT" env-default:"5432"`
	DBName   string `env:"POSTGRES_DB" env-default:"GriBD"`
	SSLMode  string `env:"POSTGRES_SSL" env-default:"disable"`
}

func New(cfg Config) (*sqlx.DB, error) {
	const op = "storage.postgres.New"

	dsn := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable host=%s port=%d", cfg.UserName, cfg.Password, cfg.DBName, cfg.Host, cfg.Port)
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if _, err := db.Conn(context.Background()); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return db, nil
}
