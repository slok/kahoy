package memory_test

import (
	"context"
	"errors"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/slok/kahoy/internal/internalerrors"
	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/storage"
	"github.com/slok/kahoy/internal/storage/memory"
)

func TestResourceRepositoryGetResource(t *testing.T) {
	tests := map[string]struct {
		repo        func() memory.ResourceRepository
		id          string
		expResource *model.Resource
		expErr      error
	}{
		"Getting a resource from an empty memory should fail.": {
			repo: func() memory.ResourceRepository {
				return memory.NewResourceRepository(nil)
			},
			id:     "test-id",
			expErr: internalerrors.ErrMissing,
		},

		"Getting a resource that is not in memory should fail.": {
			repo: func() memory.ResourceRepository {
				return memory.NewResourceRepository(map[string]model.Resource{
					"test-id2": {},
				})
			},
			id:     "test-id",
			expErr: internalerrors.ErrMissing,
		},

		"Getting a resource that is in memory should not fail and return.": {
			repo: func() memory.ResourceRepository {
				return memory.NewResourceRepository(map[string]model.Resource{
					"test-id0": {ID: "test-id0"},
					"test-id":  {ID: "test-id"},
					"test-id2": {ID: "test-id2"},
				})
			},
			id:          "test-id",
			expResource: &model.Resource{ID: "test-id"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			repo := test.repo()
			gotResource, err := repo.GetResource(context.TODO(), test.id)

			if test.expErr != nil {
				assert.True(errors.Is(err, test.expErr))
			} else if assert.NoError(err) {
				assert.Equal(test.expResource, gotResource)
			}
		})
	}
}

func TestResourceRepositoryListResources(t *testing.T) {
	tests := map[string]struct {
		repo         func() memory.ResourceRepository
		opts         storage.ResourceListOpts
		expResources *storage.ResourceList
		expErr       error
	}{
		"Listing resources from an empty memory should not fail.": {
			repo: func() memory.ResourceRepository {
				return memory.NewResourceRepository(nil)
			},
			expResources: &storage.ResourceList{
				Items: []model.Resource{},
			},
		},

		"Listing resources from memory should not fail.": {
			repo: func() memory.ResourceRepository {
				return memory.NewResourceRepository(map[string]model.Resource{
					"test-id0": {ID: "test-id0"},
					"test-id":  {ID: "test-id"},
					"test-id2": {ID: "test-id2"},
				})
			},
			expResources: &storage.ResourceList{
				Items: []model.Resource{
					{ID: "test-id0"},
					{ID: "test-id"},
					{ID: "test-id2"},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			repo := test.repo()
			gotResources, err := repo.ListResources(context.TODO(), test.opts)

			if test.expErr != nil {
				assert.True(errors.Is(err, test.expErr))
			} else if assert.NoError(err) {
				// Sort for test reliability.
				sortResourceList(test.expResources)
				sortResourceList(gotResources)
				assert.Equal(test.expResources, gotResources)
			}
		})
	}
}

func sortResourceList(l *storage.ResourceList) {
	sort.SliceStable(l.Items, func(i, j int) bool {
		return l.Items[i].ID < l.Items[j].ID
	})
}

func TestGroupRepositoryGetGroup(t *testing.T) {
	tests := map[string]struct {
		repo     func() memory.GroupRepository
		id       string
		expGroup *model.Group
		expErr   error
	}{
		"Getting a group from an empty memory should fail.": {
			repo: func() memory.GroupRepository {
				return memory.NewGroupRepository(nil)
			},
			id:     "test-id",
			expErr: internalerrors.ErrMissing,
		},
		"Getting a group that is not in memory should fail.": {
			repo: func() memory.GroupRepository {
				return memory.NewGroupRepository(map[string]model.Group{
					"test-id2": {},
				})
			},
			id:     "test-id",
			expErr: internalerrors.ErrMissing,
		},

		"Getting a group that is in memory should not fail and return.": {
			repo: func() memory.GroupRepository {
				return memory.NewGroupRepository(map[string]model.Group{
					"test-id0": {ID: "test-id0"},
					"test-id":  {ID: "test-id"},
					"test-id2": {ID: "test-id2"},
				})
			},
			id:       "test-id",
			expGroup: &model.Group{ID: "test-id"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			repo := test.repo()
			gotGroup, err := repo.GetGroup(context.TODO(), test.id)

			if test.expErr != nil {
				assert.True(errors.Is(err, test.expErr))
			} else if assert.NoError(err) {
				assert.Equal(test.expGroup, gotGroup)
			}
		})
	}
}

func TestResourceGroupListResources(t *testing.T) {
	tests := map[string]struct {
		repo      func() memory.GroupRepository
		opts      storage.GroupListOpts
		expGroups *storage.GroupList
		expErr    error
	}{
		"Listing groups from an empty memory should not fail.": {
			repo: func() memory.GroupRepository {
				return memory.NewGroupRepository(nil)
			},
			expGroups: &storage.GroupList{
				Items: []model.Group{},
			},
		},

		"Listing groups from memory should not fail.": {
			repo: func() memory.GroupRepository {
				return memory.NewGroupRepository(map[string]model.Group{
					"test-id0": {ID: "test-id0"},
					"test-id":  {ID: "test-id"},
					"test-id2": {ID: "test-id2"},
				})
			},
			expGroups: &storage.GroupList{
				Items: []model.Group{
					{ID: "test-id0"},
					{ID: "test-id"},
					{ID: "test-id2"},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			repo := test.repo()
			gotGroups, err := repo.ListGroups(context.TODO(), test.opts)

			if test.expErr != nil {
				assert.True(errors.Is(err, test.expErr))
			} else if assert.NoError(err) {
				// Sort for test reliability.
				sortGroupList(test.expGroups)
				sortGroupList(gotGroups)
				assert.Equal(test.expGroups, gotGroups)
			}
		})
	}
}

func sortGroupList(l *storage.GroupList) {
	sort.SliceStable(l.Items, func(i, j int) bool {
		return l.Items[i].ID < l.Items[j].ID
	})
}
