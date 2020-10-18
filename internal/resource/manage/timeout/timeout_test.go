package timeout_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/resource/manage"
	"github.com/slok/kahoy/internal/resource/manage/timeout"
	"github.com/stretchr/testify/require"
)

func TestTimeoutManagerApply(t *testing.T) {
	tests := map[string]struct {
		config  timeout.TimeoutManagerConfig
		manager manage.ResourceManager
		expErr  bool
	}{
		"Basic initialization should not fail.": {
			manager: testManager{},
		},

		"If Apply takes longer than timeout, apply should fail.": {
			config: timeout.TimeoutManagerConfig{
				Timeout: 1 * time.Nanosecond,
			},
			manager: testManager{},
			expErr:  true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Execute.
			test.config.Manager = test.manager
			manager, err := timeout.NewTimeoutManager(test.config)
			require.NoError(err)

			err = manager.Apply(context.Background(), []model.Resource{})

			// Check.
			if test.expErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

// testManager is a custom resource manager that handles context deadline
// exceeded and returns nil error
type testManager struct{}

// Delete provides a mock function with given fields: ctx, resources and also
// handles context done
func (n testManager) Apply(ctx context.Context, resources []model.Resource) error {
	// Handle context deadline before noop
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}

// Delete provides a mock function with given fields: ctx, resources and also
// handles context done
func (n testManager) Delete(ctx context.Context, resources []model.Resource) error {
	// Handle context deadline before noop
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}
