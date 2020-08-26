package configuration

import (
	"context"
	"fmt"

	"github.com/ghodss/yaml"

	"github.com/slok/kahoy/internal/model"
)

type jsonV1 struct {
	Version string `json:"version"`

	Groups []struct {
		ID       string `json:"id"`
		Priority *int   `json:"priority,omitempty"`
	} `json:"groups"`
}

func (j jsonV1) toModel() (*model.AppConfig, error) {
	// Map groups.
	groups := map[string]model.GroupConfig{}
	for _, g := range j.Groups {
		if g.ID == "" {
			return nil, fmt.Errorf("group id empty")
		}

		groups[g.ID] = model.GroupConfig{
			Priority: g.Priority,
		}
	}

	return &model.AppConfig{
		Groups: groups,
	}, nil
}

// NewYAMLV1Loader returns a loader that knows how to load configuration from a
// YAML string.
func NewYAMLV1Loader(data string) Loader {
	return LoaderFunc(func(ctx context.Context) (*model.AppConfig, error) {
		jv1 := &jsonV1{}
		err := yaml.Unmarshal([]byte(data), &jv1)
		if err != nil {
			return nil, fmt.Errorf("could not load unmarshal YAML configuration: %w", err)
		}

		if jv1.Version != "v1" {
			return nil, fmt.Errorf("not valid YAML configuration, required configuration version is v1")
		}

		c, err := jv1.toModel()
		if err != nil {
			return nil, fmt.Errorf("not valid YAML v1 configuration: %w", err)
		}

		return c, nil
	})
}
