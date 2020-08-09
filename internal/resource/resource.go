package resource

import (
	"context"

	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
)

// Manager knows how to manage resources on clusters.
type Manager interface {
	Apply(ctx context.Context, resources []model.Resource) error
	Delete(ctx context.Context, resources []model.Resource) error
}

type noopManager struct {
	logger log.Logger
}

// NewNoopManager returns a resource manager that noops the operations and logs them
func NewNoopManager(logger log.Logger) Manager {
	return noopManager{
		logger: logger.WithValues(log.Kv{"app-svc": "resource.NoopManager"}),
	}
}

func (n noopManager) Apply(ctx context.Context, resources []model.Resource) error {
	for _, r := range resources {
		logger := resourceLogger(n.logger, r)
		logger.Infof("apply ignored by noop manager")
	}

	return nil
}

func (n noopManager) Delete(ctx context.Context, resources []model.Resource) error {
	for _, r := range resources {
		logger := resourceLogger(n.logger, r)
		logger.Infof("delete ignored by noop manager")
	}

	return nil
}

func resourceLogger(l log.Logger, r model.Resource) log.Logger {
	return l.WithValues(log.Kv{
		"resource-id":       r.ID,
		"resource-group-id": r.GroupID,
	})
}
