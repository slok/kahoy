package kubectl_test

import (
	"context"
	"errors"
	"io/ioutil"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/resource/manage/kubectl"
	"github.com/slok/kahoy/internal/resource/manage/kubectl/kubectlmock"
	"github.com/slok/kahoy/internal/resource/manage/managemock"
)

// expCmdMatcher returns an *exec.Cmd matcher for  mock.MatchedBy.
func expCmdMatcherWithRetStderr(expArgs []string, returnStderrData string) func(cmd *exec.Cmd) bool {
	return func(cmd *exec.Cmd) bool {
		_, _ = cmd.Stderr.Write([]byte(returnStderrData))

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

func TestNamespaceEnsurerApply(t *testing.T) {
	tests := map[string]struct {
		config    kubectl.NamespaceEnsurerConfig
		resources []model.Resource
		mock      func(mrm *managemock.ResourceManager, mc *kubectlmock.CmdRunner)
		expErr    bool
	}{
		"Not having resources, should delegate to the delegated manager.": {
			resources: []model.Resource{},
			mock: func(mrm *managemock.ResourceManager, mc *kubectlmock.CmdRunner) {
				mrm.On("Apply", mock.Anything, []model.Resource{}).Once().Return(nil)
			},
		},

		"Having an error on delegated manager, should fail.": {
			resources: []model.Resource{},
			mock: func(mrm *managemock.ResourceManager, mc *kubectlmock.CmdRunner) {
				mrm.On("Apply", mock.Anything, mock.Anything).Once().Return(errors.New("whatever"))
			},
			expErr: true,
		},

		"Having resources with missing namespaces, should create the namespace once per unique namespace and delegate apply afterwards.": {
			resources: []model.Resource{
				{ID: "test1", K8sObject: newK8sObject("test1", "ns1")},
				{ID: "test2", K8sObject: newK8sObject("test2", "ns2")},
				{ID: "test3", K8sObject: newK8sObject("test3", "ns1")},
			},
			mock: func(mrm *managemock.ResourceManager, mc *kubectlmock.CmdRunner) {
				// Expect namespaces checks and creation.
				exp := expCmdMatcherWithRetStderr([]string{"kubectl", "get", "namespace", "ns1"}, `Error from server (NotFound): namespaces "ns1" not found`)
				mc.On("Run", mock.MatchedBy(exp)).Once().Return(errors.New("no namespace"))
				exp = expCmdMatcher([]string{"kubectl", "create", "namespace", "ns1"}, "")
				mc.On("Run", mock.MatchedBy(exp)).Once().Return(nil)

				exp = expCmdMatcherWithRetStderr([]string{"kubectl", "get", "namespace", "ns2"}, `Error from server (NotFound): namespaces "ns1" not found`)
				mc.On("Run", mock.MatchedBy(exp)).Once().Return(errors.New("no namespace"))
				exp = expCmdMatcher([]string{"kubectl", "create", "namespace", "ns2"}, "")
				mc.On("Run", mock.MatchedBy(exp)).Once().Return(nil)

				// Expect delegation.
				expRes := []model.Resource{
					{ID: "test1", K8sObject: newK8sObject("test1", "ns1")},
					{ID: "test2", K8sObject: newK8sObject("test2", "ns2")},
					{ID: "test3", K8sObject: newK8sObject("test3", "ns1")},
				}
				mrm.On("Apply", mock.Anything, expRes).Once().Return(nil)
			},
		},

		"Having an error while checking namespace existence should fail.": {
			resources: []model.Resource{
				{ID: "test1", K8sObject: newK8sObject("test1", "ns1")},
			},
			mock: func(mrm *managemock.ResourceManager, mc *kubectlmock.CmdRunner) {
				// Expect namespaces checks.
				exp := expCmdMatcherWithRetStderr([]string{"kubectl", "get", "namespace", "ns1"}, `other error getting the namespace`)
				mc.On("Run", mock.MatchedBy(exp)).Once().Return(errors.New("no namespace"))

				mrm.On("Apply", mock.Anything, mock.Anything).Once().Return(nil)
			},
			expErr: true,
		},

		"Having resources with present namespaces, should not create the namespaces.": {
			resources: []model.Resource{
				{ID: "test1", K8sObject: newK8sObject("test1", "ns1")},
			},
			mock: func(mrm *managemock.ResourceManager, mc *kubectlmock.CmdRunner) {
				// Expect namespaces checks.
				exp := expCmdMatcher([]string{"kubectl", "get", "namespace", "ns1"}, "")
				mc.On("Run", mock.MatchedBy(exp)).Once().Return(nil)

				mrm.On("Apply", mock.Anything, mock.Anything).Once().Return(nil)
			},
		},

		"Empty namespaces should be ignored.": {
			resources: []model.Resource{
				{ID: "test1", K8sObject: newK8sObject("test1", "")},
			},
			mock: func(mrm *managemock.ResourceManager, mc *kubectlmock.CmdRunner) {
				mrm.On("Apply", mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Mocks.
			mrm := &managemock.ResourceManager{}
			mcmd := &kubectlmock.CmdRunner{}
			test.mock(mrm, mcmd)

			// Prepare.
			test.config.Out = ioutil.Discard
			test.config.CmdRunner = mcmd
			test.config.Manager = mrm
			manager, err := kubectl.NewNamespaceEnsurer(test.config)
			require.NoError(err)

			// Execute.
			err = manager.Apply(context.TODO(), test.resources)

			// Check.
			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				mrm.AssertExpectations(t)
				mcmd.AssertExpectations(t)
			}
		})
	}
}

func TestNamespaceEnsurerDelete(t *testing.T) {
	tests := map[string]struct {
		config    kubectl.NamespaceEnsurerConfig
		resources []model.Resource
		mock      func(mrm *managemock.ResourceManager, mc *kubectlmock.CmdRunner)
		expErr    bool
	}{
		"Not having resources, should delegate to the delegated manager.": {
			resources: []model.Resource{},
			mock: func(mrm *managemock.ResourceManager, mc *kubectlmock.CmdRunner) {
				mrm.On("Delete", mock.Anything, []model.Resource{}).Once().Return(nil)
			},
		},

		"Having an error on delegated manager, should fail.": {
			resources: []model.Resource{},
			mock: func(mrm *managemock.ResourceManager, mc *kubectlmock.CmdRunner) {
				mrm.On("Delete", mock.Anything, mock.Anything).Once().Return(errors.New("whatever"))
			},
			expErr: true,
		},

		"Having resources should delegate to the delegated manager.": {
			resources: []model.Resource{
				{ID: "test1", K8sObject: newK8sObject("test1", "ns1")},
				{ID: "test2", K8sObject: newK8sObject("test2", "ns2")},
				{ID: "test3", K8sObject: newK8sObject("test3", "ns1")},
			},
			mock: func(mrm *managemock.ResourceManager, mc *kubectlmock.CmdRunner) {
				// Expect delegation.
				expRes := []model.Resource{
					{ID: "test1", K8sObject: newK8sObject("test1", "ns1")},
					{ID: "test2", K8sObject: newK8sObject("test2", "ns2")},
					{ID: "test3", K8sObject: newK8sObject("test3", "ns1")},
				}
				mrm.On("Delete", mock.Anything, expRes).Once().Return(nil)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Mocks.
			mrm := &managemock.ResourceManager{}
			mcmd := &kubectlmock.CmdRunner{}
			test.mock(mrm, mcmd)

			// Prepare.
			test.config.Out = ioutil.Discard
			test.config.CmdRunner = mcmd
			test.config.Manager = mrm
			manager, err := kubectl.NewNamespaceEnsurer(test.config)
			require.NoError(err)

			// Execute.
			err = manager.Delete(context.TODO(), test.resources)

			// Check.
			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				mrm.AssertExpectations(t)
				mcmd.AssertExpectations(t)
			}
		})
	}
}
