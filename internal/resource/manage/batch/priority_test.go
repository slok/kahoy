package batch_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/resource/manage/batch"
	"github.com/slok/kahoy/internal/resource/manage/managemock"
	"github.com/slok/kahoy/internal/storage/storagemock"
)

func TestPriorityBatchManagerApply(t *testing.T) {
	tests := map[string]struct {
		config    batch.PriorityManagerConfig
		resources []model.Resource
		mock      func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository)
		expErr    bool
	}{
		"No resources should be a noop.": {
			mock: func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository) {},
		},

		"If getting groups returns an error, it should fail.": {
			resources: []model.Resource{
				{ID: "resource1", GroupID: "group1"},
			},
			mock: func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository) {
				mgr.On("GetGroup", mock.Anything, mock.Anything).Once().Return(nil, errors.New("whatever"))
			},
			expErr: true,
		},

		"If applying resources returns an error, it should fail.": {
			resources: []model.Resource{
				{ID: "resource1", GroupID: "group1"},
			},
			mock: func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository) {
				group1 := &model.Group{ID: "group1"}
				mgr.On("GetGroup", mock.Anything, mock.Anything).Once().Return(group1, nil)
				mrm.On("Apply", mock.Anything, mock.Anything).Once().Return(errors.New("whatever"))
			},
			expErr: true,
		},

		"Resources with the same group priority and different groups, should make a single batch.": {
			resources: []model.Resource{
				{ID: "resource1", GroupID: "group1"},
				{ID: "resource2", GroupID: "group2"},
				{ID: "resource3", GroupID: "group1"},
				{ID: "resource4", GroupID: "group3"},
				{ID: "resource5", GroupID: "group2"},
			},
			mock: func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository) {
				group1 := &model.Group{ID: "group1"}
				group2 := &model.Group{ID: "group2"}
				group3 := &model.Group{ID: "group3"}
				mgr.On("GetGroup", mock.Anything, "group1").Times(2).Return(group1, nil)
				mgr.On("GetGroup", mock.Anything, "group2").Times(2).Return(group2, nil)
				mgr.On("GetGroup", mock.Anything, "group3").Once().Return(group3, nil)

				// Single batch.
				expBatch := []model.Resource{
					{ID: "resource1", GroupID: "group1"},
					{ID: "resource2", GroupID: "group2"},
					{ID: "resource3", GroupID: "group1"},
					{ID: "resource4", GroupID: "group3"},
					{ID: "resource5", GroupID: "group2"},
				}
				mrm.On("Apply", mock.Anything, expBatch).Once().Return(nil)
			},
		},

		"Resources with the different group priority should make a batch by priority and order by priority.": {
			resources: []model.Resource{
				{ID: "resource1", GroupID: "group1"},
				{ID: "resource2", GroupID: "group2"},
				{ID: "resource3", GroupID: "group1"},
				{ID: "resource4", GroupID: "group3"},
				{ID: "resource5", GroupID: "group2"},
			},
			mock: func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository) {
				group1 := &model.Group{ID: "group1", Priority: 235}
				group2 := &model.Group{ID: "group2", Priority: 42}
				group3 := &model.Group{ID: "group3", Priority: 579}
				mgr.On("GetGroup", mock.Anything, "group1").Times(2).Return(group1, nil)
				mgr.On("GetGroup", mock.Anything, "group2").Times(2).Return(group2, nil)
				mgr.On("GetGroup", mock.Anything, "group3").Once().Return(group3, nil)

				// Multiple batches.
				expBatch1 := []model.Resource{
					{ID: "resource2", GroupID: "group2"},
					{ID: "resource5", GroupID: "group2"},
				}
				mrm.On("Apply", mock.Anything, expBatch1).Once().Return(nil)

				expBatch2 := []model.Resource{
					{ID: "resource1", GroupID: "group1"},
					{ID: "resource3", GroupID: "group1"},
				}
				mrm.On("Apply", mock.Anything, expBatch2).Once().Return(nil)

				expBatch3 := []model.Resource{
					{ID: "resource4", GroupID: "group3"},
				}
				mrm.On("Apply", mock.Anything, expBatch3).Once().Return(nil)
			},
		},

		"Resources with the different group priority and priorities disabled should make a single batch.": {
			config: batch.PriorityManagerConfig{
				DisablePriorities: true,
			},
			resources: []model.Resource{
				{ID: "resource1", GroupID: "group1"},
				{ID: "resource2", GroupID: "group2"},
			},
			mock: func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository) {
				// Single batch.
				expBatch := []model.Resource{
					{ID: "resource1", GroupID: "group1"},
					{ID: "resource2", GroupID: "group2"},
				}
				mrm.On("Apply", mock.Anything, expBatch).Once().Return(nil)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Mocks.
			mrm := &managemock.ResourceManager{}
			mgr := &storagemock.GroupRepository{}
			test.mock(mrm, mgr)

			// Execute.
			test.config.Manager = mrm
			test.config.GroupRepository = mgr
			manager, err := batch.NewPriorityManager(test.config)
			require.NoError(err)

			err = manager.Apply(context.TODO(), test.resources)

			// Check.
			if test.expErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
			mrm.AssertExpectations(t)
			mgr.AssertExpectations(t)
		})
	}
}

func TestPriorityManagerDelete(t *testing.T) {
	tests := map[string]struct {
		config    batch.PriorityManagerConfig
		resources []model.Resource
		mock      func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository)
		expErr    bool
	}{
		"If applying resources returns an error, it should fail.": {
			resources: []model.Resource{
				{ID: "resource1", GroupID: "group1"},
			},
			mock: func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository) {
				mrm.On("Delete", mock.Anything, mock.Anything).Once().Return(errors.New("whatever"))
			},
			expErr: true,
		},

		"Delete should ignore priorites.": {
			resources: []model.Resource{
				{ID: "resource1", GroupID: "group1"},
				{ID: "resource2", GroupID: "group2"},
				{ID: "resource3", GroupID: "group1"},
				{ID: "resource4", GroupID: "group3"},
				{ID: "resource5", GroupID: "group2"},
			},
			mock: func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository) {
				expBatch := []model.Resource{
					{ID: "resource1", GroupID: "group1"},
					{ID: "resource2", GroupID: "group2"},
					{ID: "resource3", GroupID: "group1"},
					{ID: "resource4", GroupID: "group3"},
					{ID: "resource5", GroupID: "group2"},
				}
				mrm.On("Delete", mock.Anything, expBatch).Once().Return(nil)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Mocks.
			mrm := &managemock.ResourceManager{}
			mgr := &storagemock.GroupRepository{}
			test.mock(mrm, mgr)

			// Execute.
			test.config.Manager = mrm
			test.config.GroupRepository = mgr
			manager, err := batch.NewPriorityManager(test.config)
			require.NoError(err)

			err = manager.Delete(context.TODO(), test.resources)

			// Check.
			if test.expErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
			mrm.AssertExpectations(t)
			mgr.AssertExpectations(t)
		})
	}
}
