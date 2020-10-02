// +build integration

package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func TestKahoyApplyKubernetes(t *testing.T) {
	tests := map[string]struct {
		preCmds        []string // Will be executed before cmd.
		cmd            string
		exp            func(t *testing.T, kubeCli kubernetes.Interface)
		expStdoutParts []string // Will check stdout contains each part.
		expStderr      string
		expErr         bool
	}{
		"Applying resources should apply them on the cluster": {
			cmd: `kahoy apply --provider=kubernetes --create-namespace -n testdata/apply-all --kube-provider-id=kahoy-test --kube-provider-namespace=kahoy-integration-test`,
			exp: func(t *testing.T, cli kubernetes.Interface) {
				// Check resources exists on the cluster.
				expExists := []string{"app1", "app2", "app3", "app4", "app5"}
				for _, expRes := range expExists {
					_, err := cli.CoreV1().Services(IntegrationTestsNamespace).Get(context.TODO(), expRes, metav1.GetOptions{})
					assert.NoError(t, err)
				}
			},
			expStderr: ``,
		},

		"Having resources to delete and report enabled, it should delete them correctly from the cluster and return the report.": {
			preCmds: []string{
				// Populate the cluster with all resources.
				`kahoy apply --provider=kubernetes --create-namespace -n testdata/apply-all --kube-provider-id=kahoy-test --kube-provider-namespace=kahoy-integration-test`,
			},
			cmd: `kahoy apply --provider=kubernetes --create-namespace -n testdata/apply-some --report-path=- --kube-provider-id=kahoy-test --kube-provider-namespace=kahoy-integration-test`,
			exp: func(t *testing.T, cli kubernetes.Interface) {
				assert := assert.New(t)

				// Check resources exists on the cluster.
				expExists := []string{"app1", "app3", "app5"}
				for _, expRes := range expExists {
					_, err := cli.CoreV1().Services(IntegrationTestsNamespace).Get(context.TODO(), expRes, metav1.GetOptions{})
					assert.NoError(err)
				}

				// Check resources have been deleted from the cluster.
				expNoExists := []string{"app2", "app4"}
				for _, expRes := range expNoExists {
					_, err := cli.CoreV1().Services(IntegrationTestsNamespace).Get(context.TODO(), expRes, metav1.GetOptions{})
					assert.Error(err)
					assert.Truef(kubeerrors.IsNotFound(err), "%s should not exists the cluster, it does", expRes)
				}
			},
			expStdoutParts: []string{
				`{"id":"core/v1/Service/kahoy-integration-test/app1","group":"root","gvk":"/v1/Service","api_version":"v1","kind":"Service","namespace":"kahoy-integration-test","name":"app1"}`,
				`{"id":"core/v1/Service/kahoy-integration-test/app3","group":"root","gvk":"/v1/Service","api_version":"v1","kind":"Service","namespace":"kahoy-integration-test","name":"app3"}`,
				`{"id":"core/v1/Service/kahoy-integration-test/app5","group":"root","gvk":"/v1/Service","api_version":"v1","kind":"Service","namespace":"kahoy-integration-test","name":"app5"}`,
				`{"id":"core/v1/Service/kahoy-integration-test/app2","group":"root","gvk":"/v1/Service","api_version":"v1","kind":"Service","namespace":"kahoy-integration-test","name":"app2"}`,
				`{"id":"core/v1/Service/kahoy-integration-test/app4","group":"root","gvk":"/v1/Service","api_version":"v1","kind":"Service","namespace":"kahoy-integration-test","name":"app4"}`,
			},
			expStderr: ``,
		},

		"Having only changes, it should apply only changes correctly.": {
			preCmds: []string{
				// Populate the cluster with all resources.
				`kahoy apply --provider=kubernetes --create-namespace -n testdata/apply-all --kube-provider-id=kahoy-test --kube-provider-namespace=kahoy-integration-test`,
			},
			cmd: `kahoy apply --provider=kubernetes --create-namespace -n testdata/apply-some-changes --report-path=- --include-changes --kube-provider-id=kahoy-test --kube-provider-namespace=kahoy-integration-test`,
			exp: func(t *testing.T, cli kubernetes.Interface) {
				assert := assert.New(t)

				// Check resources exists on the cluster.
				expExists := []string{"app1", "app3", "app5"}
				for _, expRes := range expExists {
					_, err := cli.CoreV1().Services(IntegrationTestsNamespace).Get(context.TODO(), expRes, metav1.GetOptions{})
					assert.NoError(err)
				}

				// Check changes on resource that changed.
				changed, err := cli.CoreV1().Services(IntegrationTestsNamespace).Get(context.TODO(), "app5", metav1.GetOptions{})
				assert.NoError(err)
				assert.Equal(81, int(changed.Spec.Ports[0].Port))

				// Check resources have been deleted from the cluster.
				expNoExists := []string{"app2", "app4"}
				for _, expRes := range expNoExists {
					_, err := cli.CoreV1().Services(IntegrationTestsNamespace).Get(context.TODO(), expRes, metav1.GetOptions{})
					assert.Error(err)
					assert.Truef(kubeerrors.IsNotFound(err), "%s should not exists the cluster, it does", expRes)
				}
			},
			expStdoutParts: []string{
				`{"id":"core/v1/Service/kahoy-integration-test/app5","group":"root","gvk":"/v1/Service","api_version":"v1","kind":"Service","namespace":"kahoy-integration-test","name":"app5"}`,
				`{"id":"core/v1/Service/kahoy-integration-test/app2","group":"root","gvk":"/v1/Service","api_version":"v1","kind":"Service","namespace":"kahoy-integration-test","name":"app2"}`,
				`{"id":"core/v1/Service/kahoy-integration-test/app4","group":"root","gvk":"/v1/Service","api_version":"v1","kind":"Service","namespace":"kahoy-integration-test","name":"app4"}`,
			},
			expStderr: ``,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			ctx := context.Background()

			// Prepare.
			config, err := GetIntegrationConfig(ctx)
			require.NoError(err)

			cli, err := NewKubernetesClient(ctx, *config)
			require.NoError(err)

			err = CleanTestsNamespace(ctx, cli)
			require.NoError(err)

			for _, cmd := range test.preCmds {
				_, _, err := RunKahoy(ctx, *config, cmd)
				require.NoError(err)
			}

			// Execute.
			gotStdout, gotStderr, err := RunKahoy(ctx, *config, test.cmd)

			// Check.
			test.exp(t, cli)
			for _, expStdout := range test.expStdoutParts {
				assert.Contains(string(gotStdout), expStdout)
			}
			assert.Equal(test.expStderr, string(gotStderr))
			if test.expErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}
