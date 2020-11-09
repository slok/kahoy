// +build integration

package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKahoyApplyPathsDryRunChanges(t *testing.T) {
	tests := map[string]struct {
		cmd       string
		expStdout string
		expStderr string
		expErr    bool
	}{
		"Using /dev/null as old and applying all should apply all.": {
			cmd: `kahoy apply --dry-run --provider=paths -o /dev/null -n testdata/dry-run-all`,
			expStdout: `
⯈ Apply (10 resources)
├── ⯈ app2 (2 resources)
│   ├── apps/v1/Deployment/kahoy-integration-test/app2 (testdata/dry-run-all/app2/app.yaml)
│   └── core/v1/Service/kahoy-integration-test/app2 (testdata/dry-run-all/app2/svc.yaml)
├── ⯈ app2/app3 (2 resources)
│   ├── apps/v1/Deployment/kahoy-integration-test/app3 (testdata/dry-run-all/app2/app3/app3.yaml)
│   └── core/v1/Service/kahoy-integration-test/app3 (testdata/dry-run-all/app2/app3/app3.yaml)
├── ⯈ other (4 resources)
│   ├── apps/v1/Deployment/kahoy-integration-test/app4 (testdata/dry-run-all/other/app4.yaml)
│   ├── apps/v1/Deployment/kahoy-integration-test/app5 (testdata/dry-run-all/other/app5.yaml)
│   ├── core/v1/Service/kahoy-integration-test/app4 (testdata/dry-run-all/other/app4.yaml)
│   └── core/v1/Service/kahoy-integration-test/app5 (testdata/dry-run-all/other/app5.yaml)
└── ⯈ root (2 resources)
    ├── apps/v1/Deployment/kahoy-integration-test/app1 (testdata/dry-run-all/app1.yaml)
    └── core/v1/Service/kahoy-integration-test/app1 (testdata/dry-run-all/app1.yaml)

`,
		},

		"Using all as old and /dev/null as new should delete all.": {
			cmd: `kahoy apply --dry-run --provider=paths -o testdata/dry-run-all -n /dev/null`,
			expStdout: `
⯈ Delete (10 resources)
├── ⯈ app2 (2 resources)
│   ├── apps/v1/Deployment/kahoy-integration-test/app2 (testdata/dry-run-all/app2/app.yaml)
│   └── core/v1/Service/kahoy-integration-test/app2 (testdata/dry-run-all/app2/svc.yaml)
├── ⯈ app2/app3 (2 resources)
│   ├── apps/v1/Deployment/kahoy-integration-test/app3 (testdata/dry-run-all/app2/app3/app3.yaml)
│   └── core/v1/Service/kahoy-integration-test/app3 (testdata/dry-run-all/app2/app3/app3.yaml)
├── ⯈ other (4 resources)
│   ├── apps/v1/Deployment/kahoy-integration-test/app4 (testdata/dry-run-all/other/app4.yaml)
│   ├── apps/v1/Deployment/kahoy-integration-test/app5 (testdata/dry-run-all/other/app5.yaml)
│   ├── core/v1/Service/kahoy-integration-test/app4 (testdata/dry-run-all/other/app4.yaml)
│   └── core/v1/Service/kahoy-integration-test/app5 (testdata/dry-run-all/other/app5.yaml)
└── ⯈ root (2 resources)
    ├── apps/v1/Deployment/kahoy-integration-test/app1 (testdata/dry-run-all/app1.yaml)
    └── core/v1/Service/kahoy-integration-test/app1 (testdata/dry-run-all/app1.yaml)

`,
		},

		"If resource are deleted from old to new it should apply existing and delete the others.": {
			cmd: `kahoy apply --dry-run --provider=paths -o testdata/dry-run-all -n testdata/dry-run-some`,
			expStdout: `
⯈ Delete (4 resources)
├── ⯈ app2 (1 resources)
│   └── apps/v1/Deployment/kahoy-integration-test/app2 (testdata/dry-run-all/app2/app.yaml)
├── ⯈ other (2 resources)
│   ├── apps/v1/Deployment/kahoy-integration-test/app5 (testdata/dry-run-all/other/app5.yaml)
│   └── core/v1/Service/kahoy-integration-test/app5 (testdata/dry-run-all/other/app5.yaml)
└── ⯈ root (1 resources)
    └── core/v1/Service/kahoy-integration-test/app1 (testdata/dry-run-all/app1.yaml)


⯈ Apply (6 resources)
├── ⯈ app2 (1 resources)
│   └── core/v1/Service/kahoy-integration-test/app2 (testdata/dry-run-some/app2/svc.yaml)
├── ⯈ app2/app3 (2 resources)
│   ├── apps/v1/Deployment/kahoy-integration-test/app3 (testdata/dry-run-some/app2/app3/app3.yaml)
│   └── core/v1/Service/kahoy-integration-test/app3 (testdata/dry-run-some/app2/app3/app3.yaml)
├── ⯈ other (2 resources)
│   ├── apps/v1/Deployment/kahoy-integration-test/app4 (testdata/dry-run-some/other/app4.yaml)
│   └── core/v1/Service/kahoy-integration-test/app4 (testdata/dry-run-some/other/app4.yaml)
└── ⯈ root (1 resources)
    └── apps/v1/Deployment/kahoy-integration-test/app1 (testdata/dry-run-some/app1.yaml)

`,
		},

		"If resource are deleted from old to new and we only want to include changes it should apply and delete the ones changed.": {
			cmd: `kahoy apply --dry-run --provider=paths -o testdata/dry-run-all -n testdata/dry-run-some --include-changes`,
			expStdout: `
⯈ Delete (4 resources)
├── ⯈ app2 (1 resources)
│   └── apps/v1/Deployment/kahoy-integration-test/app2 (testdata/dry-run-all/app2/app.yaml)
├── ⯈ other (2 resources)
│   ├── apps/v1/Deployment/kahoy-integration-test/app5 (testdata/dry-run-all/other/app5.yaml)
│   └── core/v1/Service/kahoy-integration-test/app5 (testdata/dry-run-all/other/app5.yaml)
└── ⯈ root (1 resources)
    └── core/v1/Service/kahoy-integration-test/app1 (testdata/dry-run-all/app1.yaml)

`,
		},

		"If include changes filter is used, it should only apply changes.": {
			cmd: `kahoy apply --dry-run --provider=paths -o testdata/dry-run-some -n testdata/dry-run-some-changes --include-changes`,
			expStdout: `
⯈ Apply (2 resources)
├── ⯈ app2 (1 resources)
│   └── core/v1/Service/kahoy-integration-test/app2 (testdata/dry-run-some-changes/app2/svc.yaml)
└── ⯈ root (1 resources)
    └── apps/v1/Deployment/kahoy-integration-test/app1 (testdata/dry-run-some-changes/app1.yaml)

`,
		},

		"If include changes filter is used, it should only apply changes (using all and some).": {
			cmd: `kahoy apply --dry-run --provider=paths -o testdata/dry-run-all -n testdata/dry-run-some-changes --include-changes`,
			expStdout: `
⯈ Delete (4 resources)
├── ⯈ app2 (1 resources)
│   └── apps/v1/Deployment/kahoy-integration-test/app2 (testdata/dry-run-all/app2/app.yaml)
├── ⯈ other (2 resources)
│   ├── apps/v1/Deployment/kahoy-integration-test/app5 (testdata/dry-run-all/other/app5.yaml)
│   └── core/v1/Service/kahoy-integration-test/app5 (testdata/dry-run-all/other/app5.yaml)
└── ⯈ root (1 resources)
    └── core/v1/Service/kahoy-integration-test/app1 (testdata/dry-run-all/app1.yaml)


⯈ Apply (2 resources)
├── ⯈ app2 (1 resources)
│   └── core/v1/Service/kahoy-integration-test/app2 (testdata/dry-run-some-changes/app2/svc.yaml)
└── ⯈ root (1 resources)
    └── apps/v1/Deployment/kahoy-integration-test/app1 (testdata/dry-run-some-changes/app1.yaml)

`,
		},

		"Structure change should not affect and should not affect when detecting changes.": {
			cmd:       `kahoy apply --dry-run --provider=paths -o testdata/dry-run-all -n testdata/dry-run-all-in-one --include-changes`,
			expStdout: ``,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Prepare.
			config, err := GetIntegrationConfig(context.TODO())
			require.NoError(err)

			// Execute.
			gotStdout, gotStderr, err := RunKahoy(context.TODO(), *config, test.cmd)

			// Check.
			assert.Equal(test.expStdout, string(gotStdout))
			assert.Equal(test.expStderr, string(gotStderr))
			if test.expErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestKahoyApplyPathsDryRunFilter(t *testing.T) {
	tests := map[string]struct {
		cmd       string
		expStdout string
		expStderr string
		expErr    bool
	}{
		"Filtering by excluding path should filter resources correctly.": {
			cmd: `kahoy apply --dry-run --provider=paths -o /dev/null -n testdata/dry-run-all --fs-exclude app2\/ `,
			expStdout: `
⯈ Apply (6 resources)
├── ⯈ other (4 resources)
│   ├── apps/v1/Deployment/kahoy-integration-test/app4 (testdata/dry-run-all/other/app4.yaml)
│   ├── apps/v1/Deployment/kahoy-integration-test/app5 (testdata/dry-run-all/other/app5.yaml)
│   ├── core/v1/Service/kahoy-integration-test/app4 (testdata/dry-run-all/other/app4.yaml)
│   └── core/v1/Service/kahoy-integration-test/app5 (testdata/dry-run-all/other/app5.yaml)
└── ⯈ root (2 resources)
    ├── apps/v1/Deployment/kahoy-integration-test/app1 (testdata/dry-run-all/app1.yaml)
    └── core/v1/Service/kahoy-integration-test/app1 (testdata/dry-run-all/app1.yaml)

`,
		},

		"Filtering by including path should filter resources correctly.": {
			cmd: `kahoy apply --dry-run --provider=paths -o /dev/null -n testdata/dry-run-all --fs-include app2\/`,
			expStdout: `
⯈ Apply (4 resources)
├── ⯈ app2 (2 resources)
│   ├── apps/v1/Deployment/kahoy-integration-test/app2 (testdata/dry-run-all/app2/app.yaml)
│   └── core/v1/Service/kahoy-integration-test/app2 (testdata/dry-run-all/app2/svc.yaml)
└── ⯈ app2/app3 (2 resources)
    ├── apps/v1/Deployment/kahoy-integration-test/app3 (testdata/dry-run-all/app2/app3/app3.yaml)
    └── core/v1/Service/kahoy-integration-test/app3 (testdata/dry-run-all/app2/app3/app3.yaml)

`,
		},

		"Filtering by Kubernetes type should filter resources correctly.": {
			cmd: `kahoy apply --dry-run --provider=paths -o /dev/null -n testdata/dry-run-all --kube-exclude-type=Service`,
			expStdout: `
⯈ Apply (5 resources)
├── ⯈ app2 (1 resources)
│   └── apps/v1/Deployment/kahoy-integration-test/app2 (testdata/dry-run-all/app2/app.yaml)
├── ⯈ app2/app3 (1 resources)
│   └── apps/v1/Deployment/kahoy-integration-test/app3 (testdata/dry-run-all/app2/app3/app3.yaml)
├── ⯈ other (2 resources)
│   ├── apps/v1/Deployment/kahoy-integration-test/app4 (testdata/dry-run-all/other/app4.yaml)
│   └── apps/v1/Deployment/kahoy-integration-test/app5 (testdata/dry-run-all/other/app5.yaml)
└── ⯈ root (1 resources)
    └── apps/v1/Deployment/kahoy-integration-test/app1 (testdata/dry-run-all/app1.yaml)

`,
		},

		"Filtering by Kubernetes label should filter resources correctly.": {
			cmd: `kahoy apply --dry-run --provider=paths -o /dev/null -n testdata/dry-run-all --kube-include-label=app=app1`,
			expStdout: `
⯈ Apply (2 resources)
└── ⯈ root (2 resources)
    ├── apps/v1/Deployment/kahoy-integration-test/app1 (testdata/dry-run-all/app1.yaml)
    └── core/v1/Service/kahoy-integration-test/app1 (testdata/dry-run-all/app1.yaml)

`,
		},

		"Filtering by Kubernetes annotation should filter resources correctly.": {
			cmd: `kahoy apply --dry-run --provider=paths -o /dev/null -n testdata/dry-run-all --kube-include-annotation=app=app2`,
			expStdout: `
⯈ Apply (1 resources)
└── ⯈ app2 (1 resources)
    └── apps/v1/Deployment/kahoy-integration-test/app2 (testdata/dry-run-all/app2/app.yaml)

`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Prepare.
			config, err := GetIntegrationConfig(context.TODO())
			require.NoError(err)

			// Execute.
			gotStdout, gotStderr, err := RunKahoy(context.TODO(), *config, test.cmd)

			// Check.
			assert.Equal(test.expStdout, string(gotStdout))
			assert.Equal(test.expStderr, string(gotStderr))
			if test.expErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}
