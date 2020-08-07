package storage

import (
	"context"

	"github.com/slok/kahoy/internal/model"
)

// ResourceListOpts are the options for Resource list action on the repository.
type ResourceListOpts struct{}

// ResourceList is a list of resources.
type ResourceList struct {
	Items []model.Resource
}

// ResourceRepository knows how to retrieve Resources.
type ResourceRepository interface {
	GetResource(ctx context.Context, id string) (*model.Resource, error)
	ListResources(ctx context.Context, opts ResourceListOpts) (*ResourceList, error)
}

// GroupRepository knows to retrieve resource groups.
type GroupRepository interface {
	GetGroup(ctx context.Context, id string) (*model.Group, error)
}
