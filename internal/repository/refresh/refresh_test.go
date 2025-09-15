package refresh_test

import (
	"context"
	"log"
	"testing"
	"time"

	"auth/internal/domain/sessions"
	"auth/internal/repository/refresh"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var rdb *redis.Client
var storage *refresh.RefreshStorage

func TestMain(m *testing.M) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp").WithStartupTimeout(10 * time.Second),
	}

	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("could not start redis container: %v", err)
	}
	defer redisContainer.Terminate(ctx)

	host, _ := redisContainer.Host(ctx)
	port, _ := redisContainer.MappedPort(ctx, "6379")

	rdb = redis.NewClient(&redis.Options{
		Addr: host + ":" + port.Port(),
	})
	storage = refresh.New(rdb)

	m.Run()
}

func TestRefreshStorage_SaveGetDelete(t *testing.T) {
	ctx := context.Background()
	token := "testtoken123"
	session := sessions.RefreshSession{
		UserID:    1,
		AppID:     1,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	t.Run("save and get session", func(t *testing.T) {
		err := storage.Save(ctx, token, session)
		assert.NoError(t, err)

		got, err := storage.Get(ctx, token)
		assert.NoError(t, err)
		assert.Equal(t, session.UserID, got.UserID)
		assert.Equal(t, session.AppID, got.AppID)
	})

	t.Run("get missing token", func(t *testing.T) {
		_, err := storage.Get(ctx, "missing")
		assert.Error(t, err)
	})

	t.Run("delete token", func(t *testing.T) {
		err := storage.Delete(ctx, token)
		assert.NoError(t, err)

		_, err = storage.Get(ctx, token)
		assert.Error(t, err)
	})
}
