package plan

import (
	"context"

	"github.com/slok/kahoy/internal/model"
)

// Planner knows how to make an plan of resources.
type Planner interface {
	Plan(ctx context.Context) ([]model.ResourceGroup, error)
}
