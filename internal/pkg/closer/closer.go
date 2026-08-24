package closer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type closeFn struct {
	name string
	fn   func(context.Context) error
}

type Closer struct {
	mu    sync.Mutex
	once  sync.Once
	funcs []closeFn
	log   *slog.Logger
}

func New(log *slog.Logger) *Closer {
	return &Closer{
		log: log,
	}
}

func (c *Closer) Add(name string, fn func(context.Context) error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.funcs = append(c.funcs, closeFn{name: name, fn: fn})
}

func (c *Closer) CloseAll(ctx context.Context) error {
	var result error

	c.once.Do(func() {
		c.mu.Lock()
		funcs := c.funcs
		c.funcs = nil
		c.mu.Unlock()

		if len(funcs) == 0 {
			return
		}

		c.log.Info("started closing resources", slog.Int("count", len(funcs)))

		var errs []error

		for i := len(funcs) - 1; i >= 0; i-- {
			f := funcs[i]

			start := time.Now()
			c.log.Info("closing resource", slog.String("name", f.name))

			done := make(chan error, 1)
			go func() {
				done <- f.fn(ctx)
			}()

			select {
			case <-ctx.Done():
				err := fmt.Errorf("global timeout reached while closing %s", f.name)
				c.log.Error("global timeout", slog.String("name", f.name))
				errs = append(errs, err)
				result = errors.Join(errs...)
				return
			case <-time.After(3 * time.Second):
				err := fmt.Errorf("timeout closing %s", f.name)
				c.log.Error("resource closing timeout", slog.String("name", f.name))
				errs = append(errs, err)
			case err := <-done:
				if err != nil {
					c.log.Error("failed to close resource", slog.String("name", f.name), slog.Any("error", err))
					errs = append(errs, err)
				} else {
					c.log.Info("resource closed", slog.String("name", f.name), slog.Duration("duration", time.Since(start)))
				}
			}
		}

		c.log.Info("all resources closed")
		result = errors.Join(errs...)
	})

	return result
}
