package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	User        string        `env:"REDIS_USER" env-default:""`
	Host        string        `env:"REDIS_HOST" env-default:"localhost"`
	Port        int           `env:"REDIS_PORT" env-default:"6379"`
	DB          int           `env:"REDIS_DB" env-default:"0"`
	Password    string        `env:"REDIS_PASSWORD" env-default:""`
	MaxRetries  int           `env:"REDIS_MAX_RETRIES" env-default:"3"`
	DialTimeout time.Duration `env:"REDIS_DIAL_TIMEOUT" env-default:"5s"`
	Timeout     time.Duration `env:"REDIS_TIMEOUT" env-default:"5s"`
}

func NewClient(cfg Config) (*redis.Client, error) {
	db := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		Username:     cfg.User,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	})

	if err := db.Ping(context.Background()).Err(); err != nil {
		fmt.Printf("failed to connect to redis server: %s\n", err.Error())
		return nil, err
	}

	return db, nil
}
