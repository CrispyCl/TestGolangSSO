package refresh

import (
	"auth/internal/domain/sessions"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RefreshStorage struct {
	rdb *redis.Client
}

func New(rdb *redis.Client) *RefreshStorage {
	return &RefreshStorage{rdb: rdb}
}

func (s *RefreshStorage) Save(ctx context.Context, token string, session sessions.RefreshSession) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	key := "refresh:" + token
	return s.rdb.Set(ctx, key, data, time.Until(session.ExpiresAt)).Err()
}

func (s *RefreshStorage) Get(ctx context.Context, token string) (*sessions.RefreshSession, error) {
	key := "refresh:" + token
	data, err := s.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("refresh token not found")
	} else if err != nil {
		return nil, err
	}

	var session sessions.RefreshSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}
	return &session, nil
}

func (s *RefreshStorage) Delete(ctx context.Context, token string) error {
	key := "refresh:" + token
	return s.rdb.Del(ctx, key).Err()
}
