package app

import (
	grpcapp "auth/internal/app/grpc"
	"auth/internal/config"
	"auth/internal/repository/pg"
	"auth/internal/repository/refresh"
	"auth/internal/services/auth"
	"auth/pkg/storage/postgres"
	"auth/pkg/storage/redis"
	"log/slog"
	"time"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(log *slog.Logger, cfg config.Config) *App {
	db, err := postgres.New(cfg.Postgres)
	if err != nil {
		panic(err)
	}

	rdb, err := redis.NewClient(cfg.Redis)
	if err != nil {
		panic(err)
	}

	userRepo := pg.NewUserRepository(db)
	appRepo := pg.NewAppRepository(db)
	refreshRepo := refresh.New(rdb)

	authService := auth.New(log, userRepo, appRepo, refreshRepo, time.Duration(time.Minute*15), time.Duration(time.Hour*24*15))

	grpcApp := grpcapp.New(log, *authService, cfg.GRPCServerPort)

	return &App{GRPCServer: grpcApp}
}
