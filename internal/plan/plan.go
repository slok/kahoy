package plan

import (
	"context"

	"github.com/slok/kahoy/internal/model"
)

// ResourceState represents the state of a resource.
type ResourceState string

const (
	// ResourceStateExists represents a state where the resource should exists.
	ResourceStateExists ResourceState = "exists"
	// ResourceStateMissing represents a state where the resource should be missing.
	ResourceStateMissing ResourceState = "missing"
)

// State is the state of a plan of states.
type State struct {
	State    ResourceState
	Resource model.Resource
}

// Planner knows how to make an plan of resources based on an old group
// of resources and a new one.
type Planner interface {
	Plan(ctx context.Context, expected []model.Resource, current []model.Resource) ([]State, error)
}

type planner struct{}

// NewPlanner returns a new planner.
func NewPlanner() Planner {
	return planner{}
}

func (p planner) Plan(ctx context.Context, expected []model.Resource, current []model.Resource) ([]State, error) {
	return nil, nil
}
