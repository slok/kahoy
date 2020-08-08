package plan

import (
	"context"

	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
)

// ResourceState represents the state of a resource.
type ResourceState int

const (
	// ResourceStateUnknown represents an unknown state
	ResourceStateUnknown ResourceState = iota
	// ResourceStateExists represents a state where the resource should exists.
	ResourceStateExists
	// ResourceStateMissing represents a state where the resource should be missing.
	ResourceStateMissing
)

// State is the state of a plan of states.
type State struct {
	State    ResourceState
	Resource model.Resource
}

// Planner knows how to make an plan of resource state based on an old group
// of resources and a new one.
type Planner interface {
	Plan(ctx context.Context, expected []model.Resource, current []model.Resource) ([]State, error)
}

type planner struct {
	logger log.Logger
}

// NewPlanner returns a new planner.
func NewPlanner(logger log.Logger) Planner {
	return planner{
		logger: logger.WithValues(log.Kv{"app-svc": "plan.Planner"}),
	}
}

// Plan plans the states by comparing an expected state and the current state.
func (p planner) Plan(ctx context.Context, expected []model.Resource, current []model.Resource) ([]State, error) {
	currentIdx := indexResources(current)
	expectedIdx := indexResources(expected)

	missingQ := 0
	existsQ := 0
	states := []State{}

	// Add the ones that we know need to exist.
	for _, r := range expectedIdx {
		existsQ++
		states = append(states, State{
			State:    ResourceStateExists,
			Resource: r,
		})
	}

	// Add the ones that have been deleted.
	for id, r := range currentIdx {
		_, ok := expectedIdx[id]
		if ok {
			continue
		}

		// This resources has been deleted.
		missingQ++
		states = append(states, State{
			State:    ResourceStateMissing,
			Resource: r,
		})
	}

	p.logger.Infof("%d planned states, %d missing, %d exists", len(states), missingQ, existsQ)

	return states, nil
}

func indexResources(rs []model.Resource) map[string]model.Resource {
	index := map[string]model.Resource{}
	for _, r := range rs {
		index[r.ID] = r
	}

	return index
}
