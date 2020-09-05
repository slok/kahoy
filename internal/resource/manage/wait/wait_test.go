package wait_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/resource/manage/managemock"
	"github.com/slok/kahoy/internal/resource/manage/wait"
	"github.com/slok/kahoy/internal/resource/manage/wait/waitmock"
	"github.com/slok/kahoy/internal/storage/storagemock"
)

func TestManagerApply(t *testing.T) {
	tests := map[string]struct {
		resources []model.Resource
		mock      func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository, mtm *waitmock.TimeManager)
		expErr    bool
	}{
		"If apply has an error, it should fail.": {
			resources: []model.Resource{
				{ID: "resource1", GroupID: "group1"},
			},
			mock: func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository, mtm *waitmock.TimeManager) {
				mrm.On("Apply", mock.Anything, mock.Anything).Once().Return(errors.New("whatever"))
			},
			expErr: true,
		},

		"If getting a group, has an error, it should fail after apply.": {
			resources: []model.Resource{
				{ID: "resource1", GroupID: "group1"},
			},
			mock: func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository, mtm *waitmock.TimeManager) {
				mrm.On("Apply", mock.Anything, mock.Anything).Once().Return(nil)
				mgr.On("GetGroup", mock.Anything, mock.Anything).Once().Return(nil, errors.New("whatever"))
			},
			expErr: true,
		},

		"If any of the resources don't have a wait option, they shouldn't wait.": {
			resources: []model.Resource{
				{ID: "resource1", GroupID: "group1"},
			},
			mock: func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository, mtm *waitmock.TimeManager) {
				expResources := []model.Resource{{ID: "resource1", GroupID: "group1"}}
				mrm.On("Apply", mock.Anything, expResources).Once().Return(nil)

				group1 := &model.Group{ID: "group1"}
				mgr.On("GetGroup", mock.Anything, "group1").Once().Return(group1, nil)
			},
		},

		"If any of the resources has a wait option, they should wait.": {
			resources: []model.Resource{
				{ID: "resource1", GroupID: "group1"},
			},
			mock: func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository, mtm *waitmock.TimeManager) {
				expResources := []model.Resource{{ID: "resource1", GroupID: "group1"}}
				mrm.On("Apply", mock.Anything, expResources).Once().Return(nil)

				group1 := &model.Group{ID: "group1", Wait: model.GroupWait{Duration: 42 * time.Minute}}
				mgr.On("GetGroup", mock.Anything, "group1").Once().Return(group1, nil)

				// Expect sleep.
				mtm.On("Sleep", mock.Anything, 42*time.Minute).Once().Return()
			},
		},

		"Having multiple groups with different duration, it should wait the max duration group.": {
			resources: []model.Resource{
				{ID: "resource1", GroupID: "group1"},
				{ID: "resource2", GroupID: "group2"},
				{ID: "resource3", GroupID: "group3"},
			},
			mock: func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository, mtm *waitmock.TimeManager) {
				expResources := []model.Resource{
					{ID: "resource1", GroupID: "group1"},
					{ID: "resource2", GroupID: "group2"},
					{ID: "resource3", GroupID: "group3"},
				}
				mrm.On("Apply", mock.Anything, expResources).Once().Return(nil)

				group1 := &model.Group{ID: "group1", Wait: model.GroupWait{Duration: 42 * time.Minute}}
				group2 := &model.Group{ID: "group2", Wait: model.GroupWait{Duration: 242 * time.Minute}}
				group3 := &model.Group{ID: "group3", Wait: model.GroupWait{Duration: 142 * time.Minute}}
				mgr.On("GetGroup", mock.Anything, "group1").Once().Return(group1, nil)
				mgr.On("GetGroup", mock.Anything, "group2").Once().Return(group2, nil)
				mgr.On("GetGroup", mock.Anything, "group3").Once().Return(group3, nil)

				// Expect sleep.
				mtm.On("Sleep", mock.Anything, 242*time.Minute).Once().Return()
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
			mtm := &waitmock.TimeManager{}
			test.mock(mrm, mgr, mtm)

			// Execute.
			config := wait.ManagerConfig{
				Manager:         mrm,
				GroupRepository: mgr,
				TimeManager:     mtm,
			}
			manager, err := wait.NewManager(config)
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

func TestManagerDelete(t *testing.T) {
	tests := map[string]struct {
		resources []model.Resource
		mock      func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository, mtm *waitmock.TimeManager)
		expErr    bool
	}{
		"If delete has an error, it should fail.": {
			resources: []model.Resource{
				{ID: "resource1", GroupID: "group1"},
			},
			mock: func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository, mtm *waitmock.TimeManager) {
				mrm.On("Delete", mock.Anything, mock.Anything).Once().Return(errors.New("whatever"))
			},
			expErr: true,
		},

		"Delete should not wait.": {
			resources: []model.Resource{
				{ID: "resource1", GroupID: "group1"},
				{ID: "resource2", GroupID: "group2"},
				{ID: "resource3", GroupID: "group3"},
			},
			mock: func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository, mtm *waitmock.TimeManager) {
				expResources := []model.Resource{
					{ID: "resource1", GroupID: "group1"},
					{ID: "resource2", GroupID: "group2"},
					{ID: "resource3", GroupID: "group3"},
				}
				mrm.On("Delete", mock.Anything, expResources).Once().Return(nil)
			},
			expErr: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Mocks.
			mrm := &managemock.ResourceManager{}
			mgr := &storagemock.GroupRepository{}
			mtm := &waitmock.TimeManager{}
			test.mock(mrm, mgr, mtm)

			// Execute.
			config := wait.ManagerConfig{
				Manager:         mrm,
				GroupRepository: mgr,
				TimeManager:     mtm,
			}
			manager, err := wait.NewManager(config)
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
