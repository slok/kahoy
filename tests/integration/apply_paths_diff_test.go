package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKahoyApplyPathsDiff(t *testing.T) {
	tests := map[string]struct {
		cmd       string
		expDiff   []string
		expStderr string
		expErr    bool
	}{
		"Not having a namespace when using diff should fail.": {
			cmd: `kahoy apply --diff --provider=paths -o /dev/null -n testdata/diff-all`,
			expStderr: `error: could not apply resources correctly: could not apply batch correctly: error while running apply diff command: Error from server (NotFound): namespaces "kahoy-integration-test" not found
: exit status 2
`,
			expErr: true,
		},

		"Applying all it should create an addition diff.": {
			cmd: `kahoy apply --diff --provider=paths -o /dev/null -n testdata/diff-all --create-namespace`,
			expDiff: []string{
				// App1 diff.
				`+apiVersion: apps/v1`,
				`+kind: Deployment`,
				`+  name: app1`,

				// Svc1 diff.
				`+apiVersion: v1`,
				`+kind: Service`,
				`+  name: svc1`,
			},
		},

		"Applying invalid resources should fail.": {
			cmd: `kahoy apply --diff --provider=paths -o /dev/null -n testdata/diff-invalid --create-namespace`,
			expStderr: `error: could not apply resources correctly: could not apply batch correctly: error while running apply diff command: The Deployment "app1" is invalid: spec.replicas: Invalid value: -1: must be greater than or equal to 0
: exit status 2
`,
			expErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Prepare.
			config, err := GetIntegrationConfig(context.TODO())
			require.NoError(err)

			cli, err := NewKubernetesClient(context.TODO(), *config)
			require.NoError(err)

			err = CleanTestsNamespace(context.TODO(), cli)
			require.NoError(err)

			// Execute.
			gotStdout, gotStderr, err := RunKahoy(context.TODO(), *config, test.cmd)

			// Check.
			for _, exp := range test.expDiff {
				assert.Contains(string(gotStdout), exp)
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
