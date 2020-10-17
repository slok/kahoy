package timeout_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/resource/manage/managemock"
	"github.com/slok/kahoy/internal/resource/manage/timeout"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTimeoutManagerApply(t *testing.T) {
	tests := map[string]struct {
		config    timeout.TimeoutManagerConfig
		resources []model.Resource
		mock      func(mrm *managemock.ResourceManager)
		expErr    bool
	}{
		"Basic initialization should not fail.": {
			mock: func(mrm *managemock.ResourceManager) {
				mrm.On("Apply", mock.Anything, mock.Anything).Once().Return(nil)
			},
		},

		"If Apply takes longer than timeout, apply should fail.": {
			config: timeout.TimeoutManagerConfig{
				Timeout: 1 * time.Second,
			},
			resources: []model.Resource{
				{ID: "resource1", GroupID: "group1"},
			},
			mock: func(mrm *managemock.ResourceManager) {
				expBatch := []model.Resource{
					{ID: "resource1", GroupID: "group1"},
				}
				mrm.On("Apply", mock.Anything, expBatch).Once().Return(nil)
				// TODO(jesus.vazquez) wait few seconds so manager timeouts
			},
			expErr: false, // TODO (jesus.vazquez) switch this to true
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Mocks.
			mrm := &managemock.ResourceManager{}
			test.mock(mrm)

			// Execute.
			test.config.Manager = mrm
			manager, err := timeout.NewTimeoutManager(test.config)
			require.NoError(err)

			err = manager.Apply(context.TODO(), test.resources)

			// Check.
			if test.expErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
			mrm.AssertExpectations(t)
		})
	}
}
