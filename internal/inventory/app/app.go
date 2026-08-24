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

	"github.com/kelar1s/go-freight/internal/pkg/closer"
	"github.com/kelar1s/go-freight/internal/pkg/config"
	"github.com/kelar1s/go-freight/internal/pkg/logger"
)

type App struct {
	diContainer *diContainer
	cfg         *config.Config
	log         *slog.Logger
	httpServer  *http.Server
	closer      *closer.Closer
}

func New(cfg *config.Config, log *slog.Logger) (*App, error) {
	c := closer.New(log)

	a := &App{
		diContainer: newDIContainer(cfg, log, c),
		cfg:         cfg,
		log:         log,
		closer:      c,
	}

	if err := a.initDeps(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize dependencies: %w", err)
	}

	return a, nil
}

func (a *App) initDeps(ctx context.Context) error {
	if err := a.diContainer.InitTelemetry(ctx); err != nil {
		return err
	}

	router, err := a.diContainer.Router(ctx)
	if err != nil {
		return fmt.Errorf("failed to init router: %w", err)
	}

	a.httpServer = &http.Server{
		Addr:         a.cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  a.cfg.HTTPServer.Timeout,
		WriteTimeout: a.cfg.HTTPServer.Timeout,
		IdleTimeout:  a.cfg.HTTPServer.IdleTimeout,
	}

	return nil
}

func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a.log.Info("starting inventory service", slog.String("env", a.cfg.Env), slog.String("addr", a.cfg.HTTPServer.Address))

	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.log.Error("http server error", logger.Err(err))
		}
	}()

	<-ctx.Done()
	a.log.Info("received shutdown signal, stopping application...")
	
	stop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
		a.log.Error("failed to stop http server", logger.Err(err))
	} else {
		a.log.Info("http server gracefully stopped (no active connections)")
	}

	closerCtx, closerCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer closerCancel()

	if err := a.closer.CloseAll(closerCtx); err != nil {
		a.log.Error("failed to close resources", logger.Err(err))
	}

	return nil
}
