package kubectl_test

import (
	"context"
	"errors"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/resource/manage/kubectl"
	"github.com/slok/kahoy/internal/resource/manage/kubectl/kubectlmock"
)

func TestManagerApply(t *testing.T) {
	tests := map[string]struct {
		config    kubectl.ManagerConfig
		resources []model.Resource
		mock      func(mk *kubectlmock.K8sObjectEncoder, mc *kubectlmock.CmdRunner)
		expErr    bool
	}{
		"Not having resources, shouldn't execute anything.": {
			resources: []model.Resource{},
			mock:      func(mk *kubectlmock.K8sObjectEncoder, mc *kubectlmock.CmdRunner) {},
		},

		"Having resources should apply correctly.": {
			resources: []model.Resource{
				{Name: "test1", K8sObject: newK8sObject("test1", "ns1")},
				{Name: "test2", K8sObject: newK8sObject("test2", "ns1")},
			},
			mock: func(mk *kubectlmock.K8sObjectEncoder, mc *kubectlmock.CmdRunner) {
				expK8sResources := []model.K8sObject{
					newK8sObject("test1", "ns1"),
					newK8sObject("test2", "ns1"),
				}
				mk.On("EncodeObjects", mock.Anything, expK8sResources).Once().Return([]byte("test"), nil)

				exp := expCmdMatcher(
					[]string{"kubectl", "apply", "--force-conflicts=true", "--server-side=true", "-f", "-"},
					"test",
				)
				mc.On("Run", mock.MatchedBy(exp)).Once().Return(nil)
			},
		},

		"Having an error while encoding objects should stop the execution and fail.": {
			resources: []model.Resource{{Name: "test1"}},
			mock: func(mk *kubectlmock.K8sObjectEncoder, mc *kubectlmock.CmdRunner) {
				mk.On("EncodeObjects", mock.Anything, mock.Anything).Once().Return(nil, errors.New("whatever"))
			},
			expErr: true,
		},

		"Having an error while running the cmd should stop the execution and fail.": {
			resources: []model.Resource{{Name: "test1"}},
			mock: func(mk *kubectlmock.K8sObjectEncoder, mc *kubectlmock.CmdRunner) {
				mk.On("EncodeObjects", mock.Anything, mock.Anything).Once().Return(nil, nil)
				mc.On("Run", mock.Anything).Once().Return(errors.New("whatever"))
			},
			expErr: true,
		},

		"Having custom command should apply with the specific cmd.": {
			config: kubectl.ManagerConfig{
				KubectlCmd: "whatever",
			},
			resources: []model.Resource{{Name: "test1"}},
			mock: func(mk *kubectlmock.K8sObjectEncoder, mc *kubectlmock.CmdRunner) {
				mk.On("EncodeObjects", mock.Anything, mock.Anything).Once().Return([]byte("test"), nil)

				exp := expCmdMatcher(
					[]string{"whatever", "apply", "--force-conflicts=true", "--server-side=true", "-f", "-"},
					"test",
				)
				mc.On("Run", mock.MatchedBy(exp)).Once().Return(nil)
			},
		},

		"Having kubectl context set, should set the cmd flag.": {
			config: kubectl.ManagerConfig{
				KubeContext: "whatever",
			},
			resources: []model.Resource{{Name: "test1"}},
			mock: func(mk *kubectlmock.K8sObjectEncoder, mc *kubectlmock.CmdRunner) {
				mk.On("EncodeObjects", mock.Anything, mock.Anything).Once().Return([]byte("test"), nil)

				exp := expCmdMatcher(
					[]string{"kubectl", "apply", "--context", "whatever", "--force-conflicts=true", "--server-side=true", "-f", "-"},
					"test",
				)
				mc.On("Run", mock.MatchedBy(exp)).Once().Return(nil)
			},
		},

		"Having kubectl config set, should set the cmd flag.": {
			config: kubectl.ManagerConfig{
				KubeConfig: "whatever",
			},
			resources: []model.Resource{{Name: "test1"}},
			mock: func(mk *kubectlmock.K8sObjectEncoder, mc *kubectlmock.CmdRunner) {
				mk.On("EncodeObjects", mock.Anything, mock.Anything).Once().Return([]byte("test"), nil)

				exp := expCmdMatcher(
					[]string{"kubectl", "apply", "--kubeconfig", "whatever", "--force-conflicts=true", "--server-side=true", "-f", "-"},
					"test",
				)
				mc.On("Run", mock.MatchedBy(exp)).Once().Return(nil)
			},
		},

		"Having force disabled, should set the cmd flag.": {
			config: kubectl.ManagerConfig{
				DisableKubeForceConflicts: true,
			},
			resources: []model.Resource{{Name: "test1"}},
			mock: func(mk *kubectlmock.K8sObjectEncoder, mc *kubectlmock.CmdRunner) {
				mk.On("EncodeObjects", mock.Anything, mock.Anything).Once().Return([]byte("test"), nil)

				exp := expCmdMatcher(
					[]string{"kubectl", "apply", "--force-conflicts=false", "--server-side=true", "-f", "-"},
					"test",
				)
				mc.On("Run", mock.MatchedBy(exp)).Once().Return(nil)
			},
		},

		"Having field manager set, should set the cmd flag.": {
			config: kubectl.ManagerConfig{
				KubeFieldManager: "whatever",
			},
			resources: []model.Resource{{Name: "test1"}},
			mock: func(mk *kubectlmock.K8sObjectEncoder, mc *kubectlmock.CmdRunner) {
				mk.On("EncodeObjects", mock.Anything, mock.Anything).Once().Return([]byte("test"), nil)

				exp := expCmdMatcher(
					[]string{"kubectl", "apply", "--force-conflicts=true", "--field-manager", "whatever", "--server-side=true", "-f", "-"},
					"test",
				)
				mc.On("Run", mock.MatchedBy(exp)).Once().Return(nil)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Mocks.
			menc := &kubectlmock.K8sObjectEncoder{}
			mcmd := &kubectlmock.CmdRunner{}
			test.mock(menc, mcmd)

			// Prepare.
			test.config.Out = ioutil.Discard
			test.config.YAMLEncoder = menc
			test.config.CmdRunner = mcmd
			manager, err := kubectl.NewManager(test.config)
			require.NoError(err)

			// Execute.
			err = manager.Apply(context.TODO(), test.resources)

			// Check.
			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				menc.AssertExpectations(t)
				mcmd.AssertExpectations(t)
			}
		})
	}
}
func TestManagerDelete(t *testing.T) {
	tests := map[string]struct {
		config    kubectl.ManagerConfig
		resources []model.Resource
		mock      func(mk *kubectlmock.K8sObjectEncoder, mc *kubectlmock.CmdRunner)
		expErr    bool
	}{
		"Not having resources, shouldn't execute anything.": {
			resources: []model.Resource{},
			mock:      func(mk *kubectlmock.K8sObjectEncoder, mc *kubectlmock.CmdRunner) {},
		},

		"Having resources should delete correctly.": {
			resources: []model.Resource{
				{Name: "test1", K8sObject: newK8sObject("test1", "ns1")},
				{Name: "test2", K8sObject: newK8sObject("test2", "ns1")},
			},
			mock: func(mk *kubectlmock.K8sObjectEncoder, mc *kubectlmock.CmdRunner) {
				expK8sResources := []model.K8sObject{
					newK8sObject("test1", "ns1"),
					newK8sObject("test2", "ns1"),
				}
				mk.On("EncodeObjects", mock.Anything, expK8sResources).Once().Return([]byte("test"), nil)

				exp := expCmdMatcher(
					[]string{"kubectl", "delete", "--ignore-not-found=true", "-f", "-"},
					"test",
				)
				mc.On("Run", mock.MatchedBy(exp)).Once().Return(nil)
			},
		},

		"Having an error while encoding objects should stop the execution and fail.": {
			resources: []model.Resource{{Name: "test1"}},
			mock: func(mk *kubectlmock.K8sObjectEncoder, mc *kubectlmock.CmdRunner) {
				mk.On("EncodeObjects", mock.Anything, mock.Anything).Once().Return(nil, errors.New("whatever"))
			},
			expErr: true,
		},

		"Having an error while running the cmd should stop the execution and fail.": {
			resources: []model.Resource{{Name: "test1"}},
			mock: func(mk *kubectlmock.K8sObjectEncoder, mc *kubectlmock.CmdRunner) {
				mk.On("EncodeObjects", mock.Anything, mock.Anything).Once().Return(nil, nil)
				mc.On("Run", mock.Anything).Once().Return(errors.New("whatever"))
			},
			expErr: true,
		},

		"Having custom command should delete with the specific cmd.": {
			config: kubectl.ManagerConfig{
				KubectlCmd: "whatever",
			},
			resources: []model.Resource{{Name: "test1"}},
			mock: func(mk *kubectlmock.K8sObjectEncoder, mc *kubectlmock.CmdRunner) {
				mk.On("EncodeObjects", mock.Anything, mock.Anything).Once().Return([]byte("test"), nil)

				exp := expCmdMatcher(
					[]string{"whatever", "delete", "--ignore-not-found=true", "-f", "-"},
					"test",
				)
				mc.On("Run", mock.MatchedBy(exp)).Once().Return(nil)
			},
		},

		"Having kubectl context set, should set the cmd flag.": {
			config: kubectl.ManagerConfig{
				KubeContext: "whatever",
			},
			resources: []model.Resource{{Name: "test1"}},
			mock: func(mk *kubectlmock.K8sObjectEncoder, mc *kubectlmock.CmdRunner) {
				mk.On("EncodeObjects", mock.Anything, mock.Anything).Once().Return([]byte("test"), nil)

				exp := expCmdMatcher(
					[]string{"kubectl", "delete", "--context", "whatever", "--ignore-not-found=true", "-f", "-"},
					"test",
				)
				mc.On("Run", mock.MatchedBy(exp)).Once().Return(nil)
			},
		},

		"Having kubectl config set, should set the cmd flag.": {
			config: kubectl.ManagerConfig{
				KubeConfig: "whatever",
			},
			resources: []model.Resource{{Name: "test1"}},
			mock: func(mk *kubectlmock.K8sObjectEncoder, mc *kubectlmock.CmdRunner) {
				mk.On("EncodeObjects", mock.Anything, mock.Anything).Once().Return([]byte("test"), nil)

				exp := expCmdMatcher(
					[]string{"kubectl", "delete", "--kubeconfig", "whatever", "--ignore-not-found=true", "-f", "-"},
					"test",
				)
				mc.On("Run", mock.MatchedBy(exp)).Once().Return(nil)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Mocks.
			menc := &kubectlmock.K8sObjectEncoder{}
			mcmd := &kubectlmock.CmdRunner{}
			test.mock(menc, mcmd)

			// Prepare.
			test.config.Out = ioutil.Discard
			test.config.YAMLEncoder = menc
			test.config.CmdRunner = mcmd
			manager, err := kubectl.NewManager(test.config)
			require.NoError(err)

			// Execute.
			err = manager.Delete(context.TODO(), test.resources)

			// Check.
			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				menc.AssertExpectations(t)
				mcmd.AssertExpectations(t)
			}
		})
	}
}
