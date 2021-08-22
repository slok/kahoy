package timeout

import (
	"context"
	"fmt"
	"time"

	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/resource/manage"
)

// ManagerConfig is the configuration of the timeout resource manager.
type ManagerConfig struct {
	Timeout time.Duration
	// Manager is the original manager use resourced to apply and delete.
	Manager manage.ResourceManager
	Logger  log.Logger
}

func (c *ManagerConfig) defaults() error {
	if c.Manager == nil {
		return fmt.Errorf("manager is required")
	}

	if c.Logger == nil {
		c.Logger = log.Noop
	}
	c.Logger = c.Logger.WithValues(log.Kv{"app-svc": "manage.TimeoutManager"})

	if c.Timeout == 0 {
		c.Timeout = 5 * time.Minute
	}

	return nil
}

// NewManager wraps the application resource manager ensuring that the
// executions to either apply or delete resources has timeouts.
func NewManager(config ManagerConfig) (manage.ResourceManager, error) {
	err := config.defaults()
	if err != nil {
		return nil, fmt.Errorf("invalid timeout manager configuration: %w", err)
	}

	return timeoutManager{
		manager: config.Manager,
		logger:  config.Logger,
		timeout: config.Timeout,
	}, nil
}

type timeoutManager struct {
	manager manage.ResourceManager
	logger  log.Logger
	timeout time.Duration
}

// Apply wraps timeoutManger internal resourceManager Apply() method around an
// initialized context with a configured timeout.
func (t timeoutManager) Apply(ctx context.Context, resources []model.Resource) error {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	err := t.manager.Apply(ctx, resources)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			t.logger.Errorf("context cancelled by deadline after %s", time.Since(start).Round(time.Millisecond))
		}

		return err
	}

	return nil
}

// Apply wraps timeoutManger internal resourceManager Delete() method around an
// initialized context with a configured timeout.
func (t timeoutManager) Delete(ctx context.Context, resources []model.Resource) error {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	err := t.manager.Delete(ctx, resources)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			t.logger.Errorf("context cancelled by deadline after %s", time.Since(start).Round(time.Millisecond))
		}

		return err
	}

	return nil
}
