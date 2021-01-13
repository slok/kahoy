package configuration

import (
	"context"
	"fmt"
	"time"

	"github.com/ghodss/yaml"

	"github.com/slok/kahoy/internal/model"
)

type jsonV1 struct {
	Version string `json:"version"`
	Fs      struct {
		Exclude []string `json:"exclude"`
		Include []string `json:"include"`
	} `json:"fs"`
	Groups []jsonGroupV1 `json:"groups"`
}

type jsonGroupV1 struct {
	ID       string `json:"id"`
	Priority *int   `json:"priority,omitempty"`
	Hooks    struct {
		Pre  *jsonHookV1 `json:"pre,omitempty"`
		Post *jsonHookV1 `json:"post,omitempty"`
	} `json:"hooks"`
	Wait struct {
		Duration string `json:"duration,omitempty"` // Deprecated.
	} `json:"wait"`
}

type jsonHookV1 struct {
	Cmd     string `json:"cmd,omitempty"`
	Timeout string `json:"timeout,omitempty"`
}

func (j jsonV1) toModel() (*model.AppConfig, error) {
	// Map fs.
	fs := model.FsConfig{
		Exclude: j.Fs.Exclude,
		Include: j.Fs.Include,
	}

	// Map groups.
	groups := map[string]model.GroupConfig{}
	for _, g := range j.Groups {
		if g.ID == "" {
			return nil, fmt.Errorf("group id empty")
		}

		gm, err := g.toModel()
		if err != nil {
			return nil, fmt.Errorf("could not parse group %q: %w", g.ID, err)
		}
		groups[g.ID] = *gm
	}

	return &model.AppConfig{
		Fs:     fs,
		Groups: groups,
	}, nil
}

func (j jsonGroupV1) toModel() (*model.GroupConfig, error) {
	groupConfig := &model.GroupConfig{
		Priority: j.Priority,
	}

	// Don't allow deprecated waiting schema in configuration.
	if j.Wait.Duration != "" {
		return nil, fmt.Errorf("deprecated wait statement is being used, use `hooks` instead")
	}

	var err error
	if j.Hooks.Pre != nil {
		groupConfig.HooksConfig.Pre, err = j.Hooks.Pre.toModel()
		if err != nil {
			return nil, fmt.Errorf("invalid pre hook: %w", err)
		}
	}

	if j.Hooks.Post != nil {
		groupConfig.HooksConfig.Post, err = j.Hooks.Post.toModel()
		if err != nil {
			return nil, fmt.Errorf("invalid post hook: %w", err)
		}
	}

	return groupConfig, nil
}

func (j jsonHookV1) toModel() (*model.GroupHookConfigSpec, error) {
	if len(j.Cmd) == 0 {
		return nil, fmt.Errorf("hook command is required")
	}

	if j.Timeout == "" {
		j.Timeout = "0"
	}

	t, err := time.ParseDuration(j.Timeout)
	if err != nil {
		return nil, fmt.Errorf("invalid duration %s: %w", j.Timeout, err)
	}
	return &model.GroupHookConfigSpec{
		Cmd:     j.Cmd,
		Timeout: t,
	}, err
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
