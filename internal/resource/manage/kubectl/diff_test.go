package kubectl_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/resource/manage/kubectl"
	"github.com/slok/kahoy/internal/resource/manage/kubectl/kubectlmock"
)

// expCmdMatcher returns an *exec.Cmd matcher for  mock.MatchedBy.
func expCmdMatcher(expArgs []string, expInput string) func(cmd *exec.Cmd) bool {
	return func(cmd *exec.Cmd) bool {
		data, err := ioutil.ReadAll(cmd.Stdin)
		if err != nil {
			return false
		}
		// Write back in case further checks are made.
		cmd.Stdin = bytes.NewReader(data)

		if string(data) != expInput {
			return false
		}

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

func newK8sObject(name, ns string) model.K8sObject {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": ns,
			},
		},
	}
}

func TestDiffManagerApply(t *testing.T) {
	tests := map[string]struct {
		config    kubectl.DiffManagerConfig
		resources []model.Resource
		mock      func(mk *kubectlmock.K8sObjectEncoder, mc *kubectlmock.CmdRunner)
		expErr    bool
	}{
		"Not having resources, shouldn't execute anything.": {
			resources: []model.Resource{},
			mock:      func(mk *kubectlmock.K8sObjectEncoder, mc *kubectlmock.CmdRunner) {},
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
					[]string{"kubectl", "diff", "--force-conflicts=true", "--server-side=true", "--filename", "-"},
					"test",
				)
				mc.On("Run", mock.MatchedBy(exp)).Once().Return(nil)
			},
		},

		"Having custom command should apply with the specific cmd.": {
			config: kubectl.DiffManagerConfig{
				KubectlCmd: "whatever",
			},
			resources: []model.Resource{{Name: "test1"}},
			mock: func(mk *kubectlmock.K8sObjectEncoder, mc *kubectlmock.CmdRunner) {
				mk.On("EncodeObjects", mock.Anything, mock.Anything).Once().Return([]byte("test"), nil)

				exp := expCmdMatcher(
					[]string{"whatever", "diff", "--force-conflicts=true", "--server-side=true", "--filename", "-"},
					"test",
				)
				mc.On("Run", mock.MatchedBy(exp)).Once().Return(nil)
			},
		},

		"Having kubectl context set, should set the cmd flag.": {
			config: kubectl.DiffManagerConfig{
				KubeContext: "whatever",
			},
			resources: []model.Resource{{Name: "test1"}},
			mock: func(mk *kubectlmock.K8sObjectEncoder, mc *kubectlmock.CmdRunner) {
				mk.On("EncodeObjects", mock.Anything, mock.Anything).Once().Return([]byte("test"), nil)

				exp := expCmdMatcher(
					[]string{"kubectl", "diff", "--context", "whatever", "--force-conflicts=true", "--server-side=true", "--filename", "-"},
					"test",
				)
				mc.On("Run", mock.MatchedBy(exp)).Once().Return(nil)
			},
		},

		"Having kubectl config set, should set the cmd flag.": {
			config: kubectl.DiffManagerConfig{
				KubeConfig: "whatever",
			},
			resources: []model.Resource{{Name: "test1"}},
			mock: func(mk *kubectlmock.K8sObjectEncoder, mc *kubectlmock.CmdRunner) {
				mk.On("EncodeObjects", mock.Anything, mock.Anything).Once().Return([]byte("test"), nil)

				exp := expCmdMatcher(
					[]string{"kubectl", "diff", "--kubeconfig", "whatever", "--force-conflicts=true", "--server-side=true", "--filename", "-"},
					"test",
				)
				mc.On("Run", mock.MatchedBy(exp)).Once().Return(nil)
			},
		},

		"Having force disabled, should set the cmd flag.": {
			config: kubectl.DiffManagerConfig{
				DisableKubeForceConflicts: true,
			},
			resources: []model.Resource{{Name: "test1"}},
			mock: func(mk *kubectlmock.K8sObjectEncoder, mc *kubectlmock.CmdRunner) {
				mk.On("EncodeObjects", mock.Anything, mock.Anything).Once().Return([]byte("test"), nil)

				exp := expCmdMatcher(
					[]string{"kubectl", "diff", "--force-conflicts=false", "--server-side=true", "--filename", "-"},
					"test",
				)
				mc.On("Run", mock.MatchedBy(exp)).Once().Return(nil)
			},
		},

		"Having field manager set, should set the cmd flag.": {
			config: kubectl.DiffManagerConfig{
				KubeFieldManager: "whatever",
			},
			resources: []model.Resource{{Name: "test1"}},
			mock: func(mk *kubectlmock.K8sObjectEncoder, mc *kubectlmock.CmdRunner) {
				mk.On("EncodeObjects", mock.Anything, mock.Anything).Once().Return([]byte("test"), nil)

				exp := expCmdMatcher(
					[]string{"kubectl", "diff", "--force-conflicts=true", "--field-manager", "whatever", "--server-side=true", "--filename", "-"},
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
			test.config.YAMLDecoder = &kubectlmock.K8sObjectDecoder{}
			test.config.CmdRunner = mcmd
			manager, err := kubectl.NewDiffManager(test.config)
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

func TestDiffManagerDelete(t *testing.T) {
	tests := map[string]struct {
		config    kubectl.DiffManagerConfig
		resources []model.Resource
		mock      func(mke *kubectlmock.K8sObjectEncoder, mkd *kubectlmock.K8sObjectDecoder, mc *kubectlmock.CmdRunner, mfs *kubectlmock.FSManager)
		expErr    bool
	}{
		"Not having resources, shouldn't execute anything.": {
			resources: []model.Resource{},
			mock: func(mke *kubectlmock.K8sObjectEncoder, mkd *kubectlmock.K8sObjectDecoder, mc *kubectlmock.CmdRunner, mfs *kubectlmock.FSManager) {
			},
		},

		"Having an error while encoding objects to get the latest state from the resources in the server, should fail.": {
			resources: []model.Resource{
				{Name: "test1", K8sObject: newK8sObject("test1", "ns1")},
			},
			mock: func(mke *kubectlmock.K8sObjectEncoder, mkd *kubectlmock.K8sObjectDecoder, mc *kubectlmock.CmdRunner, mfs *kubectlmock.FSManager) {
				// Getting server state part.
				mke.On("EncodeObjects", mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("whatever"))
			},
			expErr: true,
		},

		"Having an error while running kubctl get to get the latest state from the resources in the server, should fail.": {
			resources: []model.Resource{
				{Name: "test1", K8sObject: newK8sObject("test1", "ns1")},
			},
			mock: func(mke *kubectlmock.K8sObjectEncoder, mkd *kubectlmock.K8sObjectDecoder, mc *kubectlmock.CmdRunner, mfs *kubectlmock.FSManager) {
				// Getting server state part.
				mke.On("EncodeObjects", mock.Anything, mock.Anything).Once().Return([]byte("test1"), nil)
				mc.On("Run", mock.Anything).Once().Return(fmt.Errorf("whatever"))
			},
			expErr: true,
		},

		"Having an error while decodding latest received state from the server, should fail.": {
			resources: []model.Resource{
				{Name: "test1", K8sObject: newK8sObject("test1", "ns1")},
			},
			mock: func(mke *kubectlmock.K8sObjectEncoder, mkd *kubectlmock.K8sObjectDecoder, mc *kubectlmock.CmdRunner, mfs *kubectlmock.FSManager) {
				// Getting server state part.
				mke.On("EncodeObjects", mock.Anything, mock.Anything).Once().Return([]byte("test1"), nil)
				mc.On("Run", mock.Anything).Once().Return(nil)
				mkd.On("DecodeObjects", mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("whatever"))
			},
			expErr: true,
		},

		"Having an error while creating the tmp dir should stop execution and fail.": {
			resources: []model.Resource{
				{Name: "test1", K8sObject: newK8sObject("test1", "ns1")},
			},
			mock: func(mke *kubectlmock.K8sObjectEncoder, mkd *kubectlmock.K8sObjectDecoder, mc *kubectlmock.CmdRunner, mfs *kubectlmock.FSManager) {
				// Getting server state part.
				mke.On("EncodeObjects", mock.Anything, mock.Anything).Once().Return([]byte("test1"), nil)
				mc.On("Run", mock.Anything).Once().Return(nil)
				mkd.On("DecodeObjects", mock.Anything, mock.Anything).Once().Return([]model.K8sObject{newK8sObject("test1", "ns1")}, nil)

				// Diff part.
				mfs.On("TempDir", mock.Anything, mock.Anything).Once().Return("", fmt.Errorf("whatever"))
			},
			expErr: true,
		},

		"Having an error while encoding resource should stop execution and fail.": {
			resources: []model.Resource{
				{Name: "test1", K8sObject: newK8sObject("test1", "ns1")},
			},
			mock: func(mke *kubectlmock.K8sObjectEncoder, mkd *kubectlmock.K8sObjectDecoder, mc *kubectlmock.CmdRunner, mfs *kubectlmock.FSManager) {
				// Getting server state part.
				mke.On("EncodeObjects", mock.Anything, mock.Anything).Once().Return([]byte("test1"), nil)
				mc.On("Run", mock.Anything).Once().Return(nil)
				mkd.On("DecodeObjects", mock.Anything, mock.Anything).Once().Return([]model.K8sObject{newK8sObject("test1", "ns1")}, nil)

				// Diff part.
				mfs.On("TempDir", mock.Anything, mock.Anything).Once().Return("", nil)
				mfs.On("RemoveAll", mock.Anything).Once().Return(nil)

				mke.On("EncodeObjects", mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("whatever"))
			},
			expErr: true,
		},

		"Having an error while storing encoded resource should stop execution and fail.": {
			resources: []model.Resource{
				{Name: "test1", K8sObject: newK8sObject("test1", "ns1")},
			},
			mock: func(mke *kubectlmock.K8sObjectEncoder, mkd *kubectlmock.K8sObjectDecoder, mc *kubectlmock.CmdRunner, mfs *kubectlmock.FSManager) {
				// Getting server state part.
				mke.On("EncodeObjects", mock.Anything, mock.Anything).Once().Return([]byte("test1"), nil)
				mc.On("Run", mock.Anything).Once().Return(nil)
				mkd.On("DecodeObjects", mock.Anything, mock.Anything).Once().Return([]model.K8sObject{newK8sObject("test1", "ns1")}, nil)

				// Diff part.
				mfs.On("TempDir", mock.Anything, mock.Anything).Once().Return("", nil)
				mfs.On("RemoveAll", mock.Anything).Once().Return(nil)

				mke.On("EncodeObjects", mock.Anything, mock.Anything).Once().Return([]byte("test1"), nil)
				mfs.On("WriteFile", mock.Anything, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("whatever"))
			},
			expErr: true,
		},

		"Having resources should delete correctly.": {
			resources: []model.Resource{
				{Name: "test1", K8sObject: newK8sObject("test1", "ns1")},
				{Name: "test2", K8sObject: newK8sObject("test2", "ns2")},
				{Name: "test3", K8sObject: newK8sObject("test3", "ns3")},
			},
			mock: func(mke *kubectlmock.K8sObjectEncoder, mkd *kubectlmock.K8sObjectDecoder, mc *kubectlmock.CmdRunner, mfs *kubectlmock.FSManager) {
				expK8sResources1 := newK8sObject("test1", "ns1")
				expK8sResources2 := newK8sObject("test2", "ns2")
				expK8sResources3 := newK8sObject("test3", "ns3")
				expK8sResources := []model.K8sObject{expK8sResources1, expK8sResources2, expK8sResources3}

				// Getting server state part.
				mke.On("EncodeObjects", mock.Anything, expK8sResources).Once().Return([]byte("test1"), nil)

				expCmdKubectl := expCmdMatcher(
					[]string{"kubectl", "get", "--ignore-not-found=true", "--output", "yaml", "--filename", "-"},
					"test1",
				)
				mc.On("Run", mock.MatchedBy(expCmdKubectl)).Once().Return(nil)

				// Emulate server has returned one less
				mkd.On("DecodeObjects", mock.Anything, mock.Anything).Once().Return([]model.K8sObject{expK8sResources1, expK8sResources2}, nil)

				// Diff part.
				mfs.On("TempDir", "", "KAHOY-").Once().Return("/tmp/KAHOY-42", nil)
				mfs.On("RemoveAll", "/tmp/KAHOY-42").Once().Return(nil)

				mke.On("EncodeObjects", mock.Anything, []model.K8sObject{expK8sResources1}).Once().Return([]byte("test1"), nil)
				mfs.On("WriteFile", "/tmp/KAHOY-42/apps.v1.Deployment.ns1.test1", []byte("test1"), mock.Anything).Once().Return(nil)
				expCmd1 := expCmdMatcher(
					[]string{"diff", "-u", "-N", "/tmp/KAHOY-42/apps.v1.Deployment.ns1.test1", "-"},
					"",
				)
				mc.On("Run", mock.MatchedBy(expCmd1)).Once().Return(nil)

				mke.On("EncodeObjects", mock.Anything, []model.K8sObject{expK8sResources2}).Once().Return([]byte("test2"), nil)
				mfs.On("WriteFile", "/tmp/KAHOY-42/apps.v1.Deployment.ns2.test2", []byte("test2"), mock.Anything).Once().Return(nil)
				expCmd2 := expCmdMatcher(
					[]string{"diff", "-u", "-N", "/tmp/KAHOY-42/apps.v1.Deployment.ns2.test2", "-"},
					"",
				)
				mc.On("Run", mock.MatchedBy(expCmd2)).Once().Return(nil)

			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Mocks.
			menc := &kubectlmock.K8sObjectEncoder{}
			mdec := &kubectlmock.K8sObjectDecoder{}
			mcmd := &kubectlmock.CmdRunner{}
			mfs := &kubectlmock.FSManager{}
			test.mock(menc, mdec, mcmd, mfs)

			// Prepare.
			test.config.Out = ioutil.Discard
			test.config.YAMLEncoder = menc
			test.config.YAMLDecoder = mdec
			test.config.CmdRunner = mcmd
			test.config.FSManager = mfs
			manager, err := kubectl.NewDiffManager(test.config)
			require.NoError(err)

			// Execute.
			err = manager.Delete(context.TODO(), test.resources)

			// Check.
			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				menc.AssertExpectations(t)
				mcmd.AssertExpectations(t)
				mfs.AssertExpectations(t)
			}
		})
	}
}
