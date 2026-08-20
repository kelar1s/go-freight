package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelar1s/go-freight/internal/inventory/repository"
	"github.com/kelar1s/go-freight/internal/inventory/repository/pg"
	"github.com/kelar1s/go-freight/internal/inventory/service"
	"github.com/kelar1s/go-freight/internal/inventory/transport/rest"
	"github.com/kelar1s/go-freight/internal/pkg/config"
	"github.com/kelar1s/go-freight/internal/pkg/logger"
	"github.com/kelar1s/go-freight/internal/pkg/postgres"
)

func Run() error {
	cfg := config.MustLoad()
	log := logger.Setup(cfg.Env)

	log.Info("starting inventory service", slog.String("env", cfg.Env))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := postgres.New(ctx, cfg.Database.DSN())
	if err != nil {
		return fmt.Errorf("init postgres: %w", err)
	}

	log.Info("connected to postgres")

	defer func() {
		log.Info("closing database connection")
		if err := db.Close(); err != nil {
			log.Error("close database connection", logger.Err(err))
		}
	}()

	queries := pg.New(db)
	warehouseRepo := repository.NewWarehouseRepo(queries)
	productRepo := repository.NewProductRepo(queries)

	warehouseService := service.NewWarehouseService(warehouseRepo)
	productService := service.NewProductService(productRepo)

	warehouseHandler := rest.NewWarehouseHandler(warehouseService, log)
	productHandler := rest.NewProductHandler(productService, log)

	router := rest.NewRouter(productHandler, warehouseHandler, log)

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	srvErr := make(chan error, 1)

	go func() {
		log.Info("http server is listening", slog.String("address", cfg.HTTPServer.Address))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			srvErr <- fmt.Errorf("listen and serve: %w", err)
		}
	}()

	select {
	case err := <-srvErr:
		return err
	case <-ctx.Done():
		log.Info("received shutdown signal, stopping server")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}

	log.Info("server gracefully stopped")
	return nil
}
