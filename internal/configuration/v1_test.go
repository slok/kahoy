package configuration_test

import (
	"context"
	"testing"
	"time"

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
fs:
  exclude:
    - a/test/*
    - b/*/test
  include:
    - c/test/*
    - /d/
groups:
  - id: "prometheus/crd"
    priority: 50
    wait:
      duration: 15s
`,
			expConfig: model.AppConfig{
				Fs: model.FsConfig{
					Exclude: []string{
						"a/test/*",
						"b/*/test",
					},
					Include: []string{
						"c/test/*",
						"/d/",
					},
				},
				Groups: map[string]model.GroupConfig{
					"prometheus/crd": {
						Priority: intVal(50),
						WaitConfig: &model.GroupWaitConfig{
							Duration: 15 * time.Second,
						},
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

		"A group without waiting options should be nil.": {
			data: `
version: v1
groups:
  - id: "test"
    priority: 50
`,
			expConfig: model.AppConfig{
				Groups: map[string]model.GroupConfig{
					"test": {
						Priority:   intVal(50),
						WaitConfig: nil,
					},
				},
			},
		},

		"Invalid wait duration should fail.": {
			data: `
version: v1
groups:
  - id: "test"
    priority: 50
    wait:
      duration: 15
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
