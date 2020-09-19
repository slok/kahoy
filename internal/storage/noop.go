package storage

import (
	"context"

	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
)

type noopStateRepository struct {
	logger log.Logger
}

// NewNoopStateRepository returns a new NOOP state repository
func NewNoopStateRepository(logger log.Logger) StateRepository {
	return noopStateRepository{
		logger: logger.WithValues(log.Kv{"app-svc": "storage.noopStateRepository"}),
	}
}

func (n noopStateRepository) StoreState(ctx context.Context, state model.State) error {
	n.logger.Debugf("ignoring state store by NOOP state repository")
	return nil
}
