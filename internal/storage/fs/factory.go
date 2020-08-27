package fs

import (
	"fmt"

	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
)

// RepositoriesConfig is the configuration for NewRepositories
type RepositoriesConfig struct {
	ExcludeRegex      []string
	IncludeRegex      []string
	OldPath           string
	NewPath           string
	KubernetesDecoder K8sObjectDecoder
	AppConfig         *model.AppConfig
	Logger            log.Logger
}

// NewRepositories is a factory that knows how to return two fs repositories based on common options
// at the end you will have an old FS repository and a new FS repository.
func NewRepositories(config RepositoriesConfig) (oldRepo, newRepo *Repository, err error) {
	oldRepo, err = NewRepository(RepositoryConfig{
		ExcludeRegex:      config.ExcludeRegex,
		IncludeRegex:      config.IncludeRegex,
		Path:              config.OldPath,
		KubernetesDecoder: config.KubernetesDecoder,
		AppConfig:         config.AppConfig,
		Logger: config.Logger.WithValues(log.Kv{
			"repo-state": "old",
		}),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("could not create old fs %q repository storage: %w", config.OldPath, err)
	}

	newRepo, err = NewRepository(RepositoryConfig{
		ExcludeRegex:      config.ExcludeRegex,
		IncludeRegex:      config.IncludeRegex,
		Path:              config.NewPath,
		KubernetesDecoder: config.KubernetesDecoder,
		AppConfig:         config.AppConfig,
		Logger: config.Logger.WithValues(log.Kv{
			"repo-state": "new",
		}),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("could not create new fs %q repository storage: %w", config.OldPath, err)
	}

	return oldRepo, newRepo, nil
}
