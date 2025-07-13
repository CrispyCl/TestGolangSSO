package main

import (
	"auth/internal/config"
	"auth/pkg/logger"
)

func main() {
	cfg := config.MustLoad()

	log := logger.SetupLogger(cfg.Env)
	log.Info("All is done")
}
