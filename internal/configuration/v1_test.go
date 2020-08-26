package configuration_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/slok/kahoy/internal/configuration"
	"github.com/slok/kahoy/internal/model"
)

func intVal(i int) *int {
	return &i
}

func TestYAMLV1(t *testing.T) {

	tests := map[string]struct {
		data      string
		expConfig model.AppConfig
		expErr    bool
	}{
		"Invalid YAML sould error.": {
			data:   `()`,
			expErr: true,
		},

		"Invalid config version sould error.": {
			data:   `version: v2`,
			expErr: true,
		},

		"Correct config file should be loaded correctly.": {
			data: `
version: v1
groups:
  - id: "prometheus/crd"
    priority: 50
`,
			expConfig: model.AppConfig{
				Groups: map[string]model.GroupConfig{
					"prometheus/crd": {
						Priority: intVal(50),
					},
				},
			},
		},

		"Empty group IDs can't be mapped to model.": {
			data: `
version: v1
groups:
  - id: ""
    priority: 50
`,
			expErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			loader := configuration.NewYAMLV1Loader(test.data)
			gotConfig, err := loader.Load(context.TODO())

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expConfig, *gotConfig)
			}
		})
	}
}
