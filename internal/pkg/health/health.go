package health

import (
	"context"
	"fmt"
	"slices"
	"sync"
)

type CheckFunc func(ctx context.Context) error

type check struct {
	name string
	fn   CheckFunc
}

type Checker struct {
	mu     sync.RWMutex
	checks []check
}

func New() *Checker {
	return &Checker{}
}

func (c *Checker) Add(name string, fn CheckFunc) {
	if fn == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.checks = append(c.checks, check{name: name, fn: fn})
}

func (c *Checker) CheckReady(ctx context.Context) error {
	if c == nil {
		return nil
	}

	c.mu.RLock()
	checks := slices.Clone(c.checks)
	c.mu.RUnlock()

	for _, chk := range checks {
		if err := chk.fn(ctx); err != nil {
			return fmt.Errorf("%s check failed: %w", chk.name, err)
		}
	}
	return nil
}
