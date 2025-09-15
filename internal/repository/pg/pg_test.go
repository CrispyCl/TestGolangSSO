package pg_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"auth/internal/domain/models"
	"auth/internal/repository"
	"auth/internal/repository/pg"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var db *sqlx.DB
var userRepo *pg.UserRepository
var appRepo *pg.AppRepository

func TestMain(m *testing.M) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "auth",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
	}

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("could not start container: %v", err)
	}
	defer func() {
		_ = postgresContainer.Terminate(ctx)
	}()

	host, _ := postgresContainer.Host(ctx)
	port, _ := postgresContainer.MappedPort(ctx, "5432")

	dsn := fmt.Sprintf("postgres://test:test@%s:%s/auth?sslmode=disable", host, port.Port())
	db, err = sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("could not connect to postgres: %v", err)
	}

	// Migrations
	migrator, err := migrate.New(migrationsPath(), dsn)
	if err != nil {
		log.Fatalf("could not init migrate: %v", err)
	}

	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("could not run migration: %v", err)
	}

	userRepo = pg.NewUserRepository(db)
	appRepo = pg.NewAppRepository(db)

	code := m.Run()
	os.Exit(code)
}

func TestUserRepository_CreateAndGet(t *testing.T) {
	ctx := context.Background()

	t.Run("create and get user", func(t *testing.T) {
		id, err := userRepo.Create(ctx, "test@mail.com", []byte("hash123"))
		assert.NoError(t, err)
		assert.True(t, id > 0)

		user, err := userRepo.Get(ctx, "test@mail.com")
		assert.NoError(t, err)
		assert.Equal(t, models.User{
			ID:       id,
			Email:    "test@mail.com",
			PassHash: []byte("hash123"),
		}, user)
	})

	t.Run("duplicate email", func(t *testing.T) {
		_, err := userRepo.Create(ctx, "test@mail.com", []byte("otherhash"))
		assert.ErrorIs(t, err, repository.ErrUserExists)
	})

	t.Run("user not found", func(t *testing.T) {
		_, err := userRepo.Get(ctx, "absent@mail.com")
		assert.ErrorIs(t, err, repository.ErrUserNotFound)
	})
}

func TestAppRepository_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("get existing app", func(t *testing.T) {
		var id int
		name := "test_app"
		access := "access_secret_123"
		refresh := "refresh_secret_456"

		err := db.QueryRowContext(
			ctx,
			`INSERT INTO apps (name, access_secret, refresh_secret)
			VALUES ($1, $2, $3) RETURNING id`,
			name, access, refresh,
		).Scan(&id)
		assert.NoError(t, err)

		app, err := appRepo.Get(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, models.App{
			ID:            id,
			Name:          name,
			AccessSecret:  access,
			RefreshSecret: refresh,
		}, app)
	})

	t.Run("app not found", func(t *testing.T) {
		_, err := appRepo.Get(ctx, 99999)
		assert.ErrorIs(t, err, repository.ErrAppNotFound)
	})
}

func migrationsPath() string {
	pwd, _ := os.Getwd()
	root := filepath.Join(pwd, "..", "..", "..")
	migrationsPath := filepath.Join(root, "migrations")
	migrationsPath = strings.ReplaceAll(migrationsPath, "\\", "/")

	return "file://" + migrationsPath
}
