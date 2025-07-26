package main

import (
	"flag"
	"fmt"

	"auth/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var (
		migrationsPath  string
		migrationsTable string
		direction       string
	)
	flag.StringVar(&direction, "direction", "up", "migration direction: up or down")
	flag.StringVar(&migrationsPath, "migration-path", "./migrations", "path to migrations")
	flag.StringVar(&migrationsTable, "migration-table", "migrations", "name of migrations table")

	cfg := config.MustLoad()

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&x-migrations-table=%s",
		cfg.Postgres.UserName,
		cfg.Postgres.Password,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.DBName,
		cfg.Postgres.SSLMode,
		migrationsTable,
	)

	m, err := migrate.New("file://"+migrationsPath, dsn)
	if err != nil {
		panic(err)
	}
	defer m.Close()

	switch direction {
	case "up":
		if err := m.Up(); err != nil {
			if err == migrate.ErrNoChange {
				fmt.Println("no migrations to apply")
				return
			}
			panic(err)
		}
	case "down":
		if err := m.Down(); err != nil {
			if err == migrate.ErrNoChange {
				fmt.Println("no migrations to down")
				return
			}
			panic(err)
		}
	default:
		panic("invalid direction: must be 'up' or 'down'")
	}

	fmt.Println("migrations applied successfully")
}
