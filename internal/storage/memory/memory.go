package memory

import (
	"context"
	"fmt"

	"github.com/slok/kahoy/internal/internalerrors"
	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/storage"
)

// ResourceRepository returns resources from memory.
type ResourceRepository struct {
	resources map[string]model.Resource
}

// Interface assertion.
var _ storage.ResourceRepository = ResourceRepository{}

// NewResourceRepository returns a new ResourceRepository.
func NewResourceRepository(resources map[string]model.Resource) ResourceRepository {
	if resources == nil {
		resources = map[string]model.Resource{}
	}

	return ResourceRepository{
		resources: resources,
	}
}

// GetResource satisfies storage.ResourceRepository interface.
func (r ResourceRepository) GetResource(_ context.Context, id string) (*model.Resource, error) {
	resource, ok := r.resources[id]
	if !ok {
		return nil, fmt.Errorf("%w: resource %q is missing", internalerrors.ErrMissing, id)
	}

	return &resource, nil
}

// ListResources satisfies storage.ResourceRepository interface.
func (r ResourceRepository) ListResources(_ context.Context, _ storage.ResourceListOpts) (*storage.ResourceList, error) {
	ress := make([]model.Resource, 0, len(r.resources))
	for _, res := range r.resources {
		ress = append(ress, res)
	}

	return &storage.ResourceList{
		Items: ress,
	}, nil
}

// GroupRepository returns group from memory.
type GroupRepository struct {
	groups map[string]model.Group
}

// Interface assertion.
var _ storage.GroupRepository = GroupRepository{}

// NewGroupRepository returns a new GroupRepository.
func NewGroupRepository(groups map[string]model.Group) GroupRepository {
	if groups == nil {
		groups = map[string]model.Group{}
	}

	return GroupRepository{
		groups: groups,
	}
}

// GetGroup satisfies storage.GroupRepository interface.
func (g GroupRepository) GetGroup(_ context.Context, id string) (*model.Group, error) {
	group, ok := g.groups[id]
	if !ok {
		return nil, fmt.Errorf("%w: group %q is missing", internalerrors.ErrMissing, id)
	}

	return &group, nil
}

// ListGroups satisfies storage.GroupRepository interface.
func (g GroupRepository) ListGroups(ctx context.Context, opts storage.GroupListOpts) (*storage.GroupList, error) {
	groups := make([]model.Group, 0, len(g.groups))
	for _, group := range g.groups {
		groups = append(groups, group)
	}

	return &storage.GroupList{
		Items: groups,
	}, nil
}
