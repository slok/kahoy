package wait

import (
	"context"
	"fmt"
	"time"

	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/resource/manage"
	"github.com/slok/kahoy/internal/storage"
)

// TimeManager knows how to manage time.
type TimeManager interface {
	Sleep(ctx context.Context, d time.Duration)
}

//go:generate mockery --case underscore --output waitmock --outpkg waitmock --name TimeManager

type stdTimeManager int

const stdTM = stdTimeManager(0)

var _ TimeManager = stdTM

func (stdTimeManager) Sleep(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}

// ManagerConfig is the configuration of the Wait manager.
type ManagerConfig struct {
	// Manager is the original manager used to apply and delete.
	Manager         manage.ResourceManager
	GroupRepository storage.GroupRepository
	TimeManager     TimeManager
	Logger          log.Logger
}

func (c *ManagerConfig) defaults() error {
	if c.Manager == nil {
		return fmt.Errorf("manager is required")
	}

	if c.Logger == nil {
		c.Logger = log.Noop
	}
	c.Logger = c.Logger.WithValues(log.Kv{"app-svc": "manage.WaitManager"})

	if c.GroupRepository == nil {
		return fmt.Errorf("group repository is required")
	}

	if c.TimeManager == nil {
		c.TimeManager = stdTM
	}

	return nil
}

// waitManager knows how to wait after the resources have been applied.
type waitManager struct {
	groupRepo   storage.GroupRepository
	manager     manage.ResourceManager
	timeManager TimeManager
	logger      log.Logger
}

// NewManager returns a manager that knows how to wait to resources after a correct `Apply`
func NewManager(config ManagerConfig) (manage.ResourceManager, error) {
	err := config.defaults()
	if err != nil {
		return nil, fmt.Errorf("could not create wait manager: %w", err)
	}
	return waitManager{
		groupRepo:   config.GroupRepository,
		manager:     config.Manager,
		timeManager: config.TimeManager,
		logger:      config.Logger,
	}, err
}

// Apply will apply and then wait using the most important waiting policy of all
// the applied resources.
func (w waitManager) Apply(ctx context.Context, resources []model.Resource) error {
	// Apply.
	err := w.manager.Apply(ctx, resources)
	if err != nil {
		return err
	}

	// Get the group with most important wait policy.
	var waitGroup model.Group
	for _, r := range resources {
		group, err := w.groupRepo.GetGroup(ctx, r.GroupID)
		if err != nil {
			return fmt.Errorf("could not get group %q: %w", r.GroupID, err)
		}

		// Check if this group should be the wait group.
		switch {
		// New group needs to wait more.
		case waitGroup.Wait.Duration <= group.Wait.Duration:
			waitGroup = *group
		}
	}

	// Wait after apply.
	if waitGroup.Wait.Duration > 0 {
		w.logger.WithValues(log.Kv{"wait-type": "duration", "wait-group-id": waitGroup.ID}).Infof("waiting %q after apply", waitGroup.Wait.Duration)
		w.timeManager.Sleep(ctx, waitGroup.Wait.Duration)
	}

	return nil
}

// Delete is NOOP on waiting.
func (w waitManager) Delete(ctx context.Context, resources []model.Resource) error {
	return w.manager.Delete(ctx, resources)
}
