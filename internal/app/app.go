package app

import (
	grpcapp "auth/internal/app/grpc"
	"auth/internal/config"
	"auth/internal/repository/pg"
	"auth/internal/services/auth"
	"auth/pkg/storage/postgres"
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

	userRepo := pg.NewUserRepository(db)
	appRepo := pg.NewAppRepository(db)

	authService := auth.New(log, userRepo, appRepo, time.Duration(time.Second), time.Duration(time.Second))

	grpcApp := grpcapp.New(log, *authService, cfg.GRPCServerPort)

	return &App{GRPCServer: grpcApp}
}
