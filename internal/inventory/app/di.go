package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/kelar1s/go-freight/internal/inventory/repository"
	"github.com/kelar1s/go-freight/internal/inventory/repository/pg"
	"github.com/kelar1s/go-freight/internal/inventory/service"
	"github.com/kelar1s/go-freight/internal/inventory/transport/rest"
	"github.com/kelar1s/go-freight/internal/pkg/cache"
	"github.com/kelar1s/go-freight/internal/pkg/closer"
	"github.com/kelar1s/go-freight/internal/pkg/config"
	"github.com/kelar1s/go-freight/internal/pkg/postgres"
	"github.com/kelar1s/go-freight/internal/pkg/telemetry"
)

type diContainer struct {
	cfg    *config.Config
	log    *slog.Logger
	closer *closer.Closer

	db       *sql.DB
	logCache *cache.LoggingCache

	warehouseRepo *repository.WarehouseRepo
	productRepo   *repository.ProductRepo

	warehouseService *service.WarehouseService
	productService   *service.ProductService

	warehouseHandler *rest.WarehouseHandler
	productHandler   *rest.ProductHandler
	router           http.Handler
}

func newDIContainer(cfg *config.Config, log *slog.Logger, c *closer.Closer) *diContainer {
	return &diContainer{
		cfg:    cfg,
		log:    log,
		closer: c,
	}
}

func (d *diContainer) InitTelemetry(ctx context.Context) error {
	shutdown, err := telemetry.Init(ctx, d.cfg.Telemetry.ServiceName, d.cfg.Telemetry.TempoEndpoint)
	if err != nil {
		return fmt.Errorf("failed to init telemetry: %w", err)
	}

	d.closer.Add("telemetry", func(ctx context.Context) error {
		return shutdown(ctx)
	})
	
	return nil
}

func (d *diContainer) DB(ctx context.Context) (*sql.DB, error) {
	if d.db == nil {
		db, err := postgres.New(ctx, d.cfg.Database.DSN())
		if err != nil {
			return nil, fmt.Errorf("failed to init postgres: %w", err)
		}

		d.log.Info("connected to postgres")
		d.closer.Add("postgres", func(_ context.Context) error {
			return db.Close()
		})
		d.db = db
	}
	return d.db, nil
}

func (d *diContainer) Redis(ctx context.Context) (*cache.LoggingCache, error) {
	if d.logCache == nil {
		redisCache, err := cache.NewRedisCache(
			d.cfg.Redis.Address(),
			d.cfg.Redis.Password,
			d.cfg.Redis.DB,
			d.cfg.Redis.DialTimeout,
			d.cfg.Redis.ReadTimeout,
			d.cfg.Redis.WriteTimeout,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to init redis: %w", err)
		}

		d.log.Info("connected to redis")
		d.closer.Add("redis", func(_ context.Context) error {
			return redisCache.Close()
		})

		d.logCache = cache.NewLoggingCache(*redisCache, d.log)
	}
	return d.logCache, nil
}

func (d *diContainer) WarehouseRepo(ctx context.Context) (*repository.WarehouseRepo, error) {
	if d.warehouseRepo == nil {
		db, err := d.DB(ctx)
		if err != nil {
			return nil, err
		}
		queries := pg.New(db)
		d.warehouseRepo = repository.NewWarehouseRepo(queries)
	}
	return d.warehouseRepo, nil
}

func (d *diContainer) ProductRepo(ctx context.Context) (*repository.ProductRepo, error) {
	if d.productRepo == nil {
		db, err := d.DB(ctx)
		if err != nil {
			return nil, err
		}
		d.productRepo = repository.NewProductRepo(db)
	}
	return d.productRepo, nil
}

func (d *diContainer) WarehouseService(ctx context.Context) (*service.WarehouseService, error) {
	if d.warehouseService == nil {
		repo, err := d.WarehouseRepo(ctx)
		if err != nil {
			return nil, err
		}
		
		redis, err := d.Redis(ctx)
		if err != nil {
			return nil, err
		}
		
		d.warehouseService = service.NewWarehouseService(repo, redis, d.cfg.Redis.TTL)
	}
	return d.warehouseService, nil
}

func (d *diContainer) ProductService(ctx context.Context) (*service.ProductService, error) {
	if d.productService == nil {
		repo, err := d.ProductRepo(ctx)
		if err != nil {
			return nil, err
		}
		
		redis, err := d.Redis(ctx)
		if err != nil {
			return nil, err
		}
		
		d.productService = service.NewProductService(repo, redis, d.cfg.Redis.TTL)
	}
	return d.productService, nil
}

func (d *diContainer) WarehouseHandler(ctx context.Context) (*rest.WarehouseHandler, error) {
	if d.warehouseHandler == nil {
		svc, err := d.WarehouseService(ctx)
		if err != nil {
			return nil, err
		}
		d.warehouseHandler = rest.NewWarehouseHandler(svc, d.log)
	}
	return d.warehouseHandler, nil
}

func (d *diContainer) ProductHandler(ctx context.Context) (*rest.ProductHandler, error) {
	if d.productHandler == nil {
		svc, err := d.ProductService(ctx)
		if err != nil {
			return nil, err
		}
		d.productHandler = rest.NewProductHandler(svc, d.log)
	}
	return d.productHandler, nil
}

func (d *diContainer) Router(ctx context.Context) (http.Handler, error) {
	if d.router == nil {
		ph, err := d.ProductHandler(ctx)
		if err != nil {
			return nil, err
		}
		
		wh, err := d.WarehouseHandler(ctx)
		if err != nil {
			return nil, err
		}
		
		d.router = rest.NewRouter(ph, wh, d.log)
	}
	return d.router, nil
}
