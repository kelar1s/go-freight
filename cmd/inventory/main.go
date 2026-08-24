package main

import (
	"os"

	"github.com/kelar1s/go-freight/internal/inventory/app"
	"github.com/kelar1s/go-freight/internal/pkg/config"
	"github.com/kelar1s/go-freight/internal/pkg/logger"
)

// @title           Go-Freight Inventory API
// @version         1.0
// @description     Сервис управления складами и товарами.
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	cfg := config.MustLoad()
	log := logger.Setup(cfg.Env)

	application, err := app.New(cfg, log)
	if err != nil {
		log.Error("failed to setup application", logger.Err(err))
		os.Exit(1)
	}

	if err := application.Run(); err != nil {
		log.Error("application terminated with error", logger.Err(err))
		os.Exit(1)
	}
}
