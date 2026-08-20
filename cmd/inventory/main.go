package main

import (
	"log/slog"
	"os"

	"github.com/kelar1s/go-freight/internal/inventory/app"
	"github.com/kelar1s/go-freight/internal/pkg/logger"
)

// @title           Go-Freight Inventory API
// @version         1.0
// @description     Сервис управления складами и товарами.
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	if err := app.Run(); err != nil {
		slog.Error("application terminated with error", logger.Err(err))
		os.Exit(1)
	}
}
