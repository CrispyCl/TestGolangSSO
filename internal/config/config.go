package config

import (
	"auth/pkg/storage/postgres"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Postgres postgres.Config

	Env            string        `env:"ENV" env-default:"local"`
	GRPCServerPort int           `env:"GRPC_SERVER_PORT"`
	Timeout        time.Duration `env:"SERVER_TIMEOUT" env-default:"10h"`
	MigrationsPath string
}

func MustLoad() *Config {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		panic(err)
	}

	return &cfg
}
