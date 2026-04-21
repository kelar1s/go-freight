package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelar1s/go-freight/internal/config"
	"github.com/kelar1s/go-freight/internal/inventory/handler"
	"github.com/kelar1s/go-freight/internal/inventory/repository"
	"github.com/kelar1s/go-freight/internal/inventory/repository/pg"
	"github.com/kelar1s/go-freight/internal/inventory/service"
	"github.com/kelar1s/go-freight/internal/server"
	_ "github.com/lib/pq"
)

const (
	envLocal = "local"
	envProd  = "prod"
	envDev   = "dev"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting gofreight app", slog.String("env", cfg.Env))

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Error("failed to initialize database connection", errField(err))
		os.Exit(1)
	}

	if err := db.Ping(); err != nil {
		log.Error("Failed to ping database", errField(err))
		os.Exit(1)
	}
	log.Info("Successfully connected to PostgreSQL")

	queries := pg.New(db)
	repo := repository.NewProductRepository(queries)
	svc := service.NewInventoryService(repo)
	productHandler := handler.NewProductHandler(svc, log)

	router := server.NewRouter(productHandler, log)

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("Starting HTTP server", slog.String("address", cfg.HTTPServer.Address))
		log.Info("Server is ready to handle requests")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Info("Failed to start server", errField(err))
			stop()
		}
	}()

	<-ctx.Done()
	log.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("server forced to shutdown", errField(err))
	}

	if err := db.Close(); err != nil {
		log.Error("failed to close database", errField(err))
	}

	log.Info("Server gracefully stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return log
}

func errField(err error) slog.Attr {
	return slog.String("error", err.Error())
}
