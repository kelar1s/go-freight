package cache

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/kelar1s/go-freight/internal/pkg/logger"
)

type LoggingCache struct {
	base RedisCache
	log  *slog.Logger
}

func NewLoggingCache(base RedisCache, log *slog.Logger) *LoggingCache {
	return &LoggingCache{
		base: base,
		log:  log,
	}
}

func (c *LoggingCache) Get(ctx context.Context, key string, dest any) error {
	err := c.base.Get(ctx, key, dest)
	if err != nil && !errors.Is(err, ErrCacheMiss) {
		c.log.WarnContext(ctx, "cache get error",
			slog.String("key", key),
			logger.Err(err),
		)
	}
	return err
}

func (c *LoggingCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	err := c.base.Set(ctx, key, value, ttl)
	if err != nil {
		c.log.WarnContext(ctx, "cache set error",
			slog.String("key", key),
			logger.Err(err),
		)
	}
	return err
}

func (c *LoggingCache) Delete(ctx context.Context, keys ...string) error {
	err := c.base.Delete(ctx, keys...)
	if err != nil {
		c.log.WarnContext(ctx, "cache delete error",
			slog.Any("keys", keys),
			logger.Err(err),
		)
	}
	return err
}
