package model_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/slok/kahoy/internal/model"
)

func TestAppConfigValidate(t *testing.T) {
	tests := map[string]struct {
		appConfig func() model.AppConfig
		expApp    model.AppConfig
		expErr    bool
	}{
		"When having config missing Groups, it should initialize them.": {
			appConfig: func() model.AppConfig {
				return model.AppConfig{}
			},
			expApp: model.AppConfig{
				Groups: map[string]model.GroupConfig{},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			config := test.appConfig()
			err := config.Validate(context.TODO())

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expApp, config)
			}
		})
	}
}
