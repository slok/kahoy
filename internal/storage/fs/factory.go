package fs

import (
	"context"
	"fmt"
	"io"

	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/storage"
)

// ResourceGroupRepository implements ResourceRepository and GroupRepository.
type ResourceGroupRepository interface {
	storage.ResourceRepository
	storage.GroupRepository
}

// RepositoriesConfig is the configuration for NewRepositories.
type RepositoriesConfig struct {
	Ctx               context.Context
	StdIn             io.Reader
	ExcludeRegex      []string
	IncludeRegex      []string
	OldPath           string
	NewPath           string
	KubernetesDecoder K8sObjectDecoder
	AppConfig         *model.AppConfig
	ModelFactory      *model.ResourceAndGroupFactory
	Logger            log.Logger
}

// NewRepositories is a factory that knows how to return two fs/stdin repositories based on common options
// at the end you will have an old FS repository and a new fs/stdin repository.
func NewRepositories(config RepositoriesConfig) (oldRepo, newRepo ResourceGroupRepository, err error) {
	oldRepo, err = NewRepository(RepositoryConfig{
		ExcludeRegex:      config.ExcludeRegex,
		IncludeRegex:      config.IncludeRegex,
		Path:              config.OldPath,
		KubernetesDecoder: config.KubernetesDecoder,
		AppConfig:         config.AppConfig,
		ModelFactory:      config.ModelFactory,
		Logger: config.Logger.WithValues(log.Kv{
			"repo-state": "old",
		}),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("could not create old fs %q repository storage: %w", config.OldPath, err)
	}

	// Select FS based repo or stdin.
	if config.NewPath == "-" {
		newRepo, err = NewIOReaderRepository(IOReaderRepositoryConfig{
			Ctx:               config.Ctx,
			Reader:            config.StdIn,
			KubernetesDecoder: config.KubernetesDecoder,
			AppConfig:         config.AppConfig,
			ModelFactory:      config.ModelFactory,
			Logger: config.Logger.WithValues(log.Kv{
				"repo-state": "new",
			}),
		})
		if err != nil {
			return nil, nil, fmt.Errorf("could not create new stdin repository storage: %w", err)
		}
	} else {
		newRepo, err = NewRepository(RepositoryConfig{
			ExcludeRegex:      config.ExcludeRegex,
			IncludeRegex:      config.IncludeRegex,
			Path:              config.NewPath,
			KubernetesDecoder: config.KubernetesDecoder,
			AppConfig:         config.AppConfig,
			ModelFactory:      config.ModelFactory,
			Logger: config.Logger.WithValues(log.Kv{
				"repo-state": "new",
			}),
		})
		if err != nil {
			return nil, nil, fmt.Errorf("could not create new fs %q repository storage: %w", config.NewPath, err)
		}
	}

	return oldRepo, newRepo, nil
}
