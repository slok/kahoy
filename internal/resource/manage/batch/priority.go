package batch

import (
	"context"
	"fmt"
	"sort"

	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/resource/manage"
	"github.com/slok/kahoy/internal/storage"
)

// PriorityManagerConfig is the configuration of the priority batch manager.
type PriorityManagerConfig struct {
	// DisablePriorities will disable batching by priorities.
	DisablePriorities bool
	// Manager is the original manager used to apply and delete.
	Manager         manage.ResourceManager
	GroupRepository storage.GroupRepository
	Logger          log.Logger
}

func (c *PriorityManagerConfig) defaults() error {
	if c.Manager == nil {
		return fmt.Errorf("manager is required")
	}

	if c.Logger == nil {
		c.Logger = log.Noop
	}
	c.Logger = c.Logger.WithValues(log.Kv{"app-svc": "manage.PriorityManager"})

	if c.GroupRepository == nil {
		return fmt.Errorf("group repository is required")
	}

	return nil
}

// NewPriorityManager returns a batch func that batches the resources by priority.
// If priorities enabled (default), batchManager will manage `apply` state
// resources (not on `deletes`) in sorted batches of priority, priorities
// can be set using group configuration.
func NewPriorityManager(config PriorityManagerConfig) (manage.ResourceManager, error) {
	err := config.defaults()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	var applyBatchFunc batchFunc
	if !config.DisablePriorities {
		applyBatchFunc = newPriorityBatchFunc(config.GroupRepository)
	}

	return batchManager{
		manager:         config.Manager,
		logger:          config.Logger,
		applyBatchFunc:  applyBatchFunc,
		deleteBatchFunc: nil, // Priorities not enabled on deletes.
	}, nil
}

// newPriorityBatchFunc returnsa batchFunc that returns the received resources batched in groups of the same priority ordered
// in asc (ascendant) order.
func newPriorityBatchFunc(groupRepo storage.GroupRepository) batchFunc {
	type priorityBatch struct {
		Priority  int
		Resources []model.Resource
	}

	return func(ctx context.Context, resources []model.Resource) ([]batch, error) {
		// Make batches by priority.
		batches := map[int]*priorityBatch{}
		for _, r := range resources {
			group, err := groupRepo.GetGroup(ctx, r.GroupID)
			if err != nil {
				return nil, fmt.Errorf("could not get group %q: %w", r.GroupID, err)
			}

			// Batch them.
			pg, ok := batches[group.Priority]
			if !ok {
				pg = &priorityBatch{Priority: group.Priority}
				batches[group.Priority] = pg
			}
			pg.Resources = append(pg.Resources, r)
		}

		// Sort them by priority (asc).
		orderedBatches := make([]priorityBatch, 0, len(batches))
		for _, batch := range batches {
			orderedBatches = append(orderedBatches, *batch)
		}
		sort.SliceStable(orderedBatches, func(i, j int) bool { return orderedBatches[i].Priority < orderedBatches[j].Priority })

		// Convert to batch type.
		res := make([]batch, 0, len(orderedBatches))
		for _, ob := range orderedBatches {
			res = append(res, batch{
				Metadata: map[string]interface{}{
					"priority":   ob.Priority,
					"batch-type": "priority",
				},
				Resources: ob.Resources,
			})
		}

		return res, nil
	}

}
