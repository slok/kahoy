package hook_test

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/resource/manage/hook"
	"github.com/slok/kahoy/internal/resource/manage/hook/hookmock"
	"github.com/slok/kahoy/internal/resource/manage/managemock"
	"github.com/slok/kahoy/internal/storage/storagemock"
)

type nopReader int

const nopR = nopReader(0)

func (nopReader) Read(p []byte) (n int, err error) { return len(p), nil }

// expCmdMatcher returns a matcher ready to be used with mock.MatchedBy
// to check *exec.Cmd args on mocks.
func expCmdMatcher(expArgs []string) func(cmd *exec.Cmd) bool {
	return func(cmd *exec.Cmd) bool {
		if len(expArgs) != len(cmd.Args) {
			return false
		}

		for i, arg := range expArgs {
			if arg != cmd.Args[i] {
				return false
			}
		}

		return true
	}
}

func TestManagerApply(t *testing.T) {
	tests := map[string]struct {
		resources []model.Resource
		mock      func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository, mcr *hookmock.CmdRunner)
		expErr    bool
	}{
		"If any of the resources don't have a hooks, they shouldn't execute any hook.": {
			resources: []model.Resource{
				{ID: "resource1", GroupID: "group1"},
			},
			mock: func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository, mcr *hookmock.CmdRunner) {
				expResources := []model.Resource{{ID: "resource1", GroupID: "group1"}}
				mrm.On("Apply", mock.Anything, expResources).Once().Return(nil)

				group1 := &model.Group{ID: "group1"}
				mgr.On("GetGroup", mock.Anything, "group1").Once().Return(group1, nil)
			},
		},

		"If apply has an error, it should fail.": {
			resources: []model.Resource{
				{ID: "resource1", GroupID: "group1"},
			},
			mock: func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository, mcr *hookmock.CmdRunner) {
				group1 := &model.Group{ID: "group1"}
				mgr.On("GetGroup", mock.Anything, "group1").Once().Return(group1, nil)

				mrm.On("Apply", mock.Anything, mock.Anything).Once().Return(errors.New("whatever"))
			},
			expErr: true,
		},

		"If getting a group, has an error, it should fail.": {
			resources: []model.Resource{
				{ID: "resource1", GroupID: "group1"},
			},
			mock: func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository, mcr *hookmock.CmdRunner) {
				mgr.On("GetGroup", mock.Anything, mock.Anything).Once().Return(nil, errors.New("whatever"))
			},
			expErr: true,
		},

		"If any of the resources has a pre hook, they should execute.": {
			resources: []model.Resource{
				{ID: "resource1", GroupID: "group1"},
			},
			mock: func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository, mcr *hookmock.CmdRunner) {
				group1 := &model.Group{ID: "group1", Hooks: model.GroupHooks{Pre: &model.GroupHookSpec{Cmd: []string{"cmd1", "prehook"}, Timeout: 42 * time.Minute}}}
				mgr.On("GetGroup", mock.Anything, "group1").Once().Return(group1, nil)

				// Pre hook.
				mcr.On("CombinedOutputPipe", mock.Anything).Once().Return(nopR, nil)
				expCmd := expCmdMatcher([]string{"cmd1", "prehook"})
				mcr.On("Start", mock.MatchedBy(expCmd)).Once().Return(nil)
				mcr.On("Wait", mock.MatchedBy(expCmd)).Once().Return(nil)

				expResources := []model.Resource{{ID: "resource1", GroupID: "group1"}}
				mrm.On("Apply", mock.Anything, expResources).Once().Return(nil)
			},
		},

		"If any of the resources has a post hook, they should execute.": {
			resources: []model.Resource{
				{ID: "resource1", GroupID: "group1"},
			},
			mock: func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository, mcr *hookmock.CmdRunner) {
				group1 := &model.Group{ID: "group1", Hooks: model.GroupHooks{Post: &model.GroupHookSpec{Cmd: []string{"cmd1", "posthook"}, Timeout: 42 * time.Minute}}}
				mgr.On("GetGroup", mock.Anything, "group1").Once().Return(group1, nil)

				expResources := []model.Resource{{ID: "resource1", GroupID: "group1"}}
				mrm.On("Apply", mock.Anything, expResources).Once().Return(nil)

				// Post hook.
				mcr.On("CombinedOutputPipe", mock.Anything).Once().Return(nopR, nil)
				expCmd := expCmdMatcher([]string{"cmd1", "posthook"})
				mcr.On("Start", mock.MatchedBy(expCmd)).Once().Return(nil)
				mcr.On("Wait", mock.MatchedBy(expCmd)).Once().Return(nil)
			},
		},

		"If a pre hook has an error it should fail without calling apply.": {
			resources: []model.Resource{
				{ID: "resource1", GroupID: "group1"},
			},
			mock: func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository, mcr *hookmock.CmdRunner) {
				group1 := &model.Group{ID: "group1", Hooks: model.GroupHooks{Pre: &model.GroupHookSpec{Cmd: []string{"cmd1", "prehook"}, Timeout: 42 * time.Minute}}}
				mgr.On("GetGroup", mock.Anything, "group1").Once().Return(group1, nil)

				// Pre hook.
				mcr.On("CombinedOutputPipe", mock.Anything).Once().Return(nopR, nil)
				expCmd := expCmdMatcher([]string{"cmd1", "prehook"})
				mcr.On("Start", mock.MatchedBy(expCmd)).Once().Return(nil)
				mcr.On("Wait", mock.MatchedBy(expCmd)).Once().Return(errors.New("whatever"))
			},
			expErr: true,
		},

		"If a pre hook has an error it should fail.": {
			resources: []model.Resource{
				{ID: "resource1", GroupID: "group1"},
			},
			mock: func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository, mcr *hookmock.CmdRunner) {
				group1 := &model.Group{ID: "group1", Hooks: model.GroupHooks{Post: &model.GroupHookSpec{Cmd: []string{"cmd1", "posthook"}, Timeout: 42 * time.Minute}}}
				mgr.On("GetGroup", mock.Anything, "group1").Once().Return(group1, nil)

				expResources := []model.Resource{{ID: "resource1", GroupID: "group1"}}
				mrm.On("Apply", mock.Anything, expResources).Once().Return(nil)

				// Post hook.
				mcr.On("CombinedOutputPipe", mock.Anything).Once().Return(nopR, nil)
				expCmd := expCmdMatcher([]string{"cmd1", "posthook"})
				mcr.On("Start", mock.MatchedBy(expCmd)).Once().Return(nil)
				mcr.On("Wait", mock.MatchedBy(expCmd)).Once().Return(fmt.Errorf("whatever"))
			},
			expErr: true,
		},

		"If any of the resources has a pre and post hook, they should execute.": {
			resources: []model.Resource{
				{ID: "resource1", GroupID: "group1"},
			},
			mock: func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository, mcr *hookmock.CmdRunner) {
				group1 := &model.Group{ID: "group1", Hooks: model.GroupHooks{
					Pre:  &model.GroupHookSpec{Cmd: []string{"cmd1", "prehook"}, Timeout: 42 * time.Minute},
					Post: &model.GroupHookSpec{Cmd: []string{"cmd2", "posthook"}, Timeout: 42 * time.Minute},
				}}
				mgr.On("GetGroup", mock.Anything, "group1").Once().Return(group1, nil)

				// Pre hook.
				mcr.On("CombinedOutputPipe", mock.Anything).Once().Return(nopR, nil)
				expCmd := expCmdMatcher([]string{"cmd1", "prehook"})
				mcr.On("Start", mock.MatchedBy(expCmd)).Once().Return(nil)
				mcr.On("Wait", mock.MatchedBy(expCmd)).Once().Return(nil)

				expResources := []model.Resource{{ID: "resource1", GroupID: "group1"}}
				mrm.On("Apply", mock.Anything, expResources).Once().Return(nil)

				// Post hook.
				mcr.On("CombinedOutputPipe", mock.Anything).Once().Return(nopR, nil)
				expCmd = expCmdMatcher([]string{"cmd2", "posthook"})
				mcr.On("Start", mock.MatchedBy(expCmd)).Once().Return(nil)
				mcr.On("Wait", mock.MatchedBy(expCmd)).Once().Return(nil)
			},
		},

		"Having multiple groups with different hooks, it should execute all.": {
			resources: []model.Resource{
				{ID: "resource1", GroupID: "group1"},
				{ID: "resource2", GroupID: "group2"},
				{ID: "resource3", GroupID: "group3"},
			},
			mock: func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository, mcr *hookmock.CmdRunner) {
				group1 := &model.Group{ID: "group1", Hooks: model.GroupHooks{
					Pre:  &model.GroupHookSpec{Cmd: []string{"cmd1", "prehook"}, Timeout: 1 * time.Minute},
					Post: &model.GroupHookSpec{Cmd: []string{"cmd2", "posthook"}, Timeout: 2 * time.Minute},
				}}
				group2 := &model.Group{ID: "group2", Hooks: model.GroupHooks{
					Pre: &model.GroupHookSpec{Cmd: []string{"cmd3", "prehook"}, Timeout: 3 * time.Minute},
				}}
				group3 := &model.Group{ID: "group3", Hooks: model.GroupHooks{
					Post: &model.GroupHookSpec{Cmd: []string{"cmd4", "posthook"}, Timeout: 4 * time.Minute},
				}}
				mgr.On("GetGroup", mock.Anything, "group1").Once().Return(group1, nil)
				mgr.On("GetGroup", mock.Anything, "group2").Once().Return(group2, nil)
				mgr.On("GetGroup", mock.Anything, "group3").Once().Return(group3, nil)

				// Pre hooks.
				mcr.On("CombinedOutputPipe", mock.Anything).Once().Return(nopR, nil)
				expCmd := expCmdMatcher([]string{"cmd1", "prehook"})
				mcr.On("Start", mock.MatchedBy(expCmd)).Once().Return(nil)
				mcr.On("Wait", mock.MatchedBy(expCmd)).Once().Return(nil)

				mcr.On("CombinedOutputPipe", mock.Anything).Once().Return(nopR, nil)
				expCmd = expCmdMatcher([]string{"cmd3", "prehook"})
				mcr.On("Start", mock.MatchedBy(expCmd)).Once().Return(nil)
				mcr.On("Wait", mock.MatchedBy(expCmd)).Once().Return(nil)

				// Execution.
				expResources := []model.Resource{
					{ID: "resource1", GroupID: "group1"},
					{ID: "resource2", GroupID: "group2"},
					{ID: "resource3", GroupID: "group3"},
				}
				mrm.On("Apply", mock.Anything, expResources).Once().Return(nil)

				// Post hooks.
				mcr.On("CombinedOutputPipe", mock.Anything).Once().Return(nopR, nil)
				expCmd = expCmdMatcher([]string{"cmd2", "posthook"})
				mcr.On("Start", mock.MatchedBy(expCmd)).Once().Return(nil)
				mcr.On("Wait", mock.MatchedBy(expCmd)).Once().Return(nil)

				mcr.On("CombinedOutputPipe", mock.Anything).Once().Return(nopR, nil)
				expCmd = expCmdMatcher([]string{"cmd4", "posthook"})
				mcr.On("Start", mock.MatchedBy(expCmd)).Once().Return(nil)
				mcr.On("Wait", mock.MatchedBy(expCmd)).Once().Return(nil)
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
			mcr := &hookmock.CmdRunner{}
			test.mock(mrm, mgr, mcr)

			// Execute.
			config := hook.ManagerConfig{
				Manager:         mrm,
				GroupRepository: mgr,
				CmdRunner:       mcr,
			}
			manager, err := hook.NewManager(config)
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
			mcr.AssertExpectations(t)
		})
	}
}

func TestManagerDelete(t *testing.T) {
	tests := map[string]struct {
		resources []model.Resource
		mock      func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository, mcr *hookmock.CmdRunner)
		expErr    bool
	}{
		"If delete has an error, it should fail.": {
			resources: []model.Resource{
				{ID: "resource1", GroupID: "group1"},
			},
			mock: func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository, mcr *hookmock.CmdRunner) {
				mrm.On("Delete", mock.Anything, mock.Anything).Once().Return(errors.New("whatever"))
			},
			expErr: true,
		},

		"Delete should not execute hooks.": {
			resources: []model.Resource{
				{ID: "resource1", GroupID: "group1"},
				{ID: "resource2", GroupID: "group2"},
				{ID: "resource3", GroupID: "group3"},
			},
			mock: func(mrm *managemock.ResourceManager, mgr *storagemock.GroupRepository, mcr *hookmock.CmdRunner) {
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
			mcr := &hookmock.CmdRunner{}
			test.mock(mrm, mgr, mcr)

			// Execute.
			config := hook.ManagerConfig{
				Manager:         mrm,
				GroupRepository: mgr,
				CmdRunner:       mcr,
			}
			manager, err := hook.NewManager(config)
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
