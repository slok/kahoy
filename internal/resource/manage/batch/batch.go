package batch

import (
	"context"
	"fmt"

	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/resource/manage"
)

type batch struct {
	Metadata  map[string]interface{}
	Resources []model.Resource
}
type batchFunc = func(ctx context.Context, resources []model.Resource) ([]batch, error)

// batchManager is a generic batching manager that knows how to batch resources and apply them
// using `applyBatchFunc` and `deleteBatchFunc`.
//
// Normally this batch manager is used internally to create different batching managers.
type batchManager struct {
	manager         manage.ResourceManager
	logger          log.Logger
	applyBatchFunc  batchFunc
	deleteBatchFunc batchFunc
}

func (b batchManager) Apply(ctx context.Context, resources []model.Resource) error {
	if b.applyBatchFunc == nil {
		return b.manager.Apply(ctx, resources)
	}

	batches, err := b.applyBatchFunc(ctx, resources)
	if err != nil {
		return fmt.Errorf("could not batch resources: %w", err)
	}

	totalBatches := len(batches)
	for i, batch := range batches {
		b.logger.WithValues(log.Kv(batch.Metadata)).Infof("applying batch %d of %d", i+1, totalBatches)
		err := b.manager.Apply(ctx, batch.Resources)
		if err != nil {
			return fmt.Errorf("could not apply batch correctly: %w", err)
		}
	}

	return nil
}

func (b batchManager) Delete(ctx context.Context, resources []model.Resource) error {
	if b.deleteBatchFunc == nil {
		return b.manager.Delete(ctx, resources)
	}

	batches, err := b.deleteBatchFunc(ctx, resources)
	if err != nil {
		return fmt.Errorf("could not batch resources: %w", err)
	}

	totalBatches := len(batches)
	for i, batch := range batches {
		b.logger.WithValues(log.Kv(batch.Metadata)).Infof("deleting batch %d of %d", i+1, totalBatches)
		err := b.manager.Delete(ctx, batch.Resources)
		if err != nil {
			return fmt.Errorf("could not delete batch correctly: %w", err)
		}
	}

	return nil
}
