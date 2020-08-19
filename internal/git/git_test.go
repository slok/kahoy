package git_test

import (
	"strings"
	"testing"

	"github.com/slok/kahoy/internal/git"
	"github.com/stretchr/testify/assert"
)

func TestDiffNameOnlyToFSInclude(t *testing.T) {
	tests := map[string]struct {
		diff        string
		expIncludes []string
	}{
		"No lines should return empty includes.": {
			diff:        "",
			expIncludes: []string{},
		},

		"Having multiple lines should return the include regexes.": {
			diff: `
README.md
manifests/grafana/grafana-dashboards/grafana-dashboards-dev.yaml
manifests/grafana/grafana-dashboards/grafana-dashboards-house.yaml
manifests/grafana/grafana-dashboards/grafana-dashboards-infra.yaml
manifests/grafana/grafana-dashboards/grafana-dashboards-kubernetes.yaml
manifests/grafana/grafana-dashboards/grafana-dashboards-provision.yaml
manifests/grafana/grafana.yaml
manifests/dex/dex-secret.yaml
manifests/dex/dex.yaml
`,
			expIncludes: []string{
				`.*\/README.md$`,
				`.*\/manifests\/grafana\/grafana-dashboards\/grafana-dashboards-dev.yaml$`,
				`.*\/manifests\/grafana\/grafana-dashboards\/grafana-dashboards-house.yaml$`,
				`.*\/manifests\/grafana\/grafana-dashboards\/grafana-dashboards-infra.yaml$`,
				`.*\/manifests\/grafana\/grafana-dashboards\/grafana-dashboards-kubernetes.yaml$`,
				`.*\/manifests\/grafana\/grafana-dashboards\/grafana-dashboards-provision.yaml$`,
				`.*\/manifests\/grafana\/grafana.yaml$`,
				`.*\/manifests\/dex\/dex-secret.yaml$`,
				`.*\/manifests\/dex\/dex.yaml$`,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			diff := strings.NewReader(test.diff)
			gotIncludes := git.DiffNameOnlyToFSInclude(diff)

			assert.Equal(test.expIncludes, gotIncludes)
		})
	}
}
