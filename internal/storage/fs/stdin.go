package fs

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/slok/kahoy/internal/internalerrors"
	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/storage"
	storagememory "github.com/slok/kahoy/internal/storage/memory"
)

// IOReaderRepository returns resources from io.Reader.
type IOReaderRepository struct {
	k8sDecoder   K8sObjectDecoder
	logger       log.Logger
	rootGroupID  string
	appConfig    model.AppConfig
	modelFactory *model.ResourceAndGroupFactory

	resourceMemoryRepo storagememory.ResourceRepository
	groupMemoryRepo    storagememory.GroupRepository
}

// Interface assertion.
var (
	_ storage.ResourceRepository = &Repository{}
	_ storage.GroupRepository    = &Repository{}
)

// IOReaderRepositoryConfig is the configuration of IOReaderRepository.
type IOReaderRepositoryConfig struct {
	Ctx               context.Context
	LoadTimeout       time.Duration
	Reader            io.Reader
	KubernetesDecoder K8sObjectDecoder
	RootGroupID       string
	Logger            log.Logger
	AppConfig         *model.AppConfig
	ModelFactory      *model.ResourceAndGroupFactory
}

func (c *IOReaderRepositoryConfig) defaults() error {
	if c.Ctx == nil {
		c.Ctx = context.Background()
	}

	if c.LoadTimeout == 0 {
		// Enough time to get data from stdin. We just want to avoid getting stuck forever
		// waiting for something that will not come.
		c.LoadTimeout = 10 * time.Second
	}

	if c.Reader == nil {
		return fmt.Errorf("reader is required")
	}

	if c.KubernetesDecoder == nil {
		return fmt.Errorf("kubernetes object loader is required")
	}

	if c.Logger == nil {
		c.Logger = log.Noop
	}
	c.Logger = c.Logger.WithValues(log.Kv{
		"app-svc": "fs.IOReaderRepository",
	})

	if c.ModelFactory == nil {
		return fmt.Errorf("resource and group model factory is required")
	}

	if c.RootGroupID == "" {
		c.RootGroupID = "root"
	}

	if c.AppConfig == nil {
		return fmt.Errorf("app configuration is required")
	}

	return nil
}

// NewIOReaderRepository returns a new NewIOReaderRepository.
// it will load all the resource and groups from the received io.Reader.
// the loaded resources will be loaded into a memory repository.
func NewIOReaderRepository(config IOReaderRepositoryConfig) (*IOReaderRepository, error) {
	err := config.defaults()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	r := &IOReaderRepository{
		k8sDecoder:   config.KubernetesDecoder,
		logger:       config.Logger,
		rootGroupID:  config.RootGroupID,
		appConfig:    *config.AppConfig,
		modelFactory: config.ModelFactory,
	}

	ctx, cancel := context.WithTimeout(config.Ctx, config.LoadTimeout)
	defer cancel()
	err = r.load(ctx, config.Reader)
	if err != nil {
		return nil, fmt.Errorf("could not load repository from reader: %w", err)
	}

	return r, nil
}

func (i *IOReaderRepository) load(ctx context.Context, r io.Reader) error {
	data, err := i.readData(ctx, r)
	if err != nil {
		return fmt.Errorf("could not read data: %w", err)
	}

	objs, err := i.k8sDecoder.DecodeObjects(context.Background(), data)
	if err != nil {
		return fmt.Errorf("could not load kubernetes objects: %w", err)
	}

	resources := map[string]model.Resource{}
	for _, obj := range objs {

		resource, err := i.modelFactory.NewResource(obj, i.rootGroupID, "stdin")
		if err != nil {
			return fmt.Errorf("could not create model resource: %w", err)
		}

		// Check if we already have this resource.
		_, ok := resources[resource.ID]
		if ok {
			return fmt.Errorf("%w: resource collision with %s", internalerrors.ErrNotValid, resource.ID)
		}

		// Store the resource.
		resources[resource.ID] = *resource
	}

	// Stdin only has the root group.
	rootGroupConfig := i.appConfig.Groups[i.rootGroupID]
	groups := map[string]model.Group{
		i.rootGroupID: i.modelFactory.NewGroup(i.rootGroupID, "", rootGroupConfig),
	}

	i.resourceMemoryRepo = storagememory.NewResourceRepository(resources)
	i.groupMemoryRepo = storagememory.NewGroupRepository(groups)

	return nil
}

func (i *IOReaderRepository) readData(ctx context.Context, r io.Reader) ([]byte, error) {
	var data []byte
	var err error

	c := make(chan error)
	go func() {
		data, err = ioutil.ReadAll(r)
		if err != nil {
			c <- fmt.Errorf("could not read data from io.Reader: %w", err)
		}
		c <- nil
	}()

	// Wait for read.
	select {
	case <-ctx.Done():
		err := fmt.Errorf("repository could not read data from io.Reader, context done")
		if ctxErr := ctx.Err(); ctxErr != nil {
			err = fmt.Errorf("%s: %w", err, ctxErr)
		}
		return nil, err
	case err = <-c:
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

// GetResource satisfies storage.ResourceRepository interface.
func (i *IOReaderRepository) GetResource(ctx context.Context, id string) (*model.Resource, error) {
	return i.resourceMemoryRepo.GetResource(ctx, id)
}

// ListResources satisfies storage.ResourceRepository interface.
func (i *IOReaderRepository) ListResources(ctx context.Context, opts storage.ResourceListOpts) (*storage.ResourceList, error) {
	return i.resourceMemoryRepo.ListResources(ctx, opts)
}

// GetGroup satisfies storage.GroupRepository interface.
func (i *IOReaderRepository) GetGroup(ctx context.Context, id string) (*model.Group, error) {
	return i.groupMemoryRepo.GetGroup(ctx, id)
}

// ListGroups satisfies storage.GroupRepository interface.
func (i *IOReaderRepository) ListGroups(ctx context.Context, opts storage.GroupListOpts) (*storage.GroupList, error) {
	return i.groupMemoryRepo.ListGroups(ctx, opts)
}
