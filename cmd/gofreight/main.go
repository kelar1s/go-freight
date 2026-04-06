package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelar1s/go-freight/internal/config"
	"github.com/kelar1s/go-freight/internal/products/handler"
	"github.com/kelar1s/go-freight/internal/products/repository"
	"github.com/kelar1s/go-freight/internal/products/repository/pg"
	"github.com/kelar1s/go-freight/internal/products/service"
	"github.com/kelar1s/go-freight/internal/server"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.MustLoad()
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
		log.Fatalf("Failed to initialize database connection: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Successfully connected to PostgreSQL")

	queries := pg.New(db)
	repo := repository.NewProductRepository(queries)
	svc := service.NewProductService(repo)
	productHandler := handler.NewProductHandler(svc)

	router := server.NewRouter(productHandler)

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
		log.Printf("Starting HTTP server on %s\n", cfg.HTTPServer.Address)
		log.Println("Server is ready to handle requests")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Failed to start server: %v", err)
			stop()
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	if err := db.Close(); err != nil {
		log.Printf("Error closing DB: %v", err)
	}

	log.Println("Server gracefully stopped")
}
