package main

import (
	"log/slog"
	"os"

	"github.com/kelar1s/go-freight/internal/inventory/app"
	"github.com/kelar1s/go-freight/internal/pkg/logger"
	_ "github.com/lib/pq"
)

func main() {
	if err := app.Run(); err != nil {
		slog.Error("application terminated with error", logger.Err(err))
		os.Exit(1)
	}
}
