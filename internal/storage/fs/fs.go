package fs

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/slok/kahoy/internal/internalerrors"
	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/storage"
	storagememory "github.com/slok/kahoy/internal/storage/memory"
)

type stdFSManager struct {
}

func (stdFSManager) Walk(root string, walkFn filepath.WalkFunc) error {
	return filepath.Walk(root, walkFn)
}

func (stdFSManager) ReadFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func (stdFSManager) Abs(path string) (string, error) {
	return filepath.Abs(path)
}

// K8sObjectDecoder knows how to decode Kubernetes object from manifests.
type K8sObjectDecoder interface {
	DecodeObjects(ctx context.Context, raw []byte) ([]model.K8sObject, error)
}

//go:generate mockery --case underscore --output fsmock --outpkg fsmock --name K8sObjectDecoder

// FileSystemManager knows how to manage file system.
type FileSystemManager interface {
	Walk(root string, walkFn filepath.WalkFunc) error
	ReadFile(path string) ([]byte, error)
	Abs(path string) (string, error)
}

//go:generate mockery --case underscore --output fsmock --outpkg fsmock --name FileSystemManager

// Repository returns resources from the file system.
type Repository struct {
	k8sDecoder    K8sObjectDecoder
	fsManager     FileSystemManager
	logger        log.Logger
	excludeRegex  []*regexp.Regexp
	includeRegex  []*regexp.Regexp
	defaultIgnore bool
	rootGroupID   string
	appConfig     model.AppConfig
	modelFactory  *model.ResourceAndGroupFactory

	resourceMemoryRepo storagememory.ResourceRepository
	groupMemoryRepo    storagememory.GroupRepository
}

// Interface assertion.
var (
	_ storage.ResourceRepository = &Repository{}
	_ storage.GroupRepository    = &Repository{}
)

// RepositoryConfig is the configuration of ResourceRepository.
type RepositoryConfig struct {
	ExcludeRegex      []string
	IncludeRegex      []string
	Path              string
	FSManager         FileSystemManager
	KubernetesDecoder K8sObjectDecoder
	RootGroupID       string
	Logger            log.Logger
	AppConfig         *model.AppConfig
	ModelFactory      *model.ResourceAndGroupFactory

	// Internal.
	compiledExcludeRegex []*regexp.Regexp
	compiledIncludeRegex []*regexp.Regexp
}

func (c *RepositoryConfig) defaults() error {
	if c.Path == "" {
		return fmt.Errorf("path is required")
	}

	if c.KubernetesDecoder == nil {
		return fmt.Errorf("kubernetes object loader is required")
	}

	if c.FSManager == nil {
		c.FSManager = stdFSManager{}
	}

	if c.Logger == nil {
		c.Logger = log.Noop
	}
	c.Logger = c.Logger.WithValues(log.Kv{
		"app-svc":   "fs.Repository",
		"root-path": c.Path,
	})

	// Clean path (Important for Group Id).
	c.Path = filepath.Clean(c.Path)

	if c.RootGroupID == "" {
		c.RootGroupID = "root"
	}

	if c.ModelFactory == nil {
		return fmt.Errorf("resource and group model factory is required")
	}

	// Compile regex.
	for _, r := range c.ExcludeRegex {
		if r == "" {
			continue
		}
		cr, err := regexp.Compile(r)
		if err != nil {
			return fmt.Errorf("could not compile %q regex: %w", r, err)
		}
		c.compiledExcludeRegex = append(c.compiledExcludeRegex, cr)
	}

	if c.AppConfig == nil {
		return fmt.Errorf("app configuration is required")
	}

	for _, r := range c.IncludeRegex {
		if r == "" {
			continue
		}
		cr, err := regexp.Compile(r)
		if err != nil {
			return fmt.Errorf("could not compile %q regex: %w", r, err)
		}
		c.compiledIncludeRegex = append(c.compiledIncludeRegex, cr)
	}

	return nil
}

// NewRepository returns a new Repository.
func NewRepository(config RepositoryConfig) (*Repository, error) {
	err := config.defaults()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	r := &Repository{
		k8sDecoder:    config.KubernetesDecoder,
		fsManager:     config.FSManager,
		logger:        config.Logger,
		rootGroupID:   config.RootGroupID,
		appConfig:     *config.AppConfig,
		modelFactory:  config.ModelFactory,
		excludeRegex:  config.compiledExcludeRegex,
		includeRegex:  config.compiledIncludeRegex,
		defaultIgnore: len(config.compiledIncludeRegex) > 0, // If we have any include rule, by default we ignore.
	}

	err = r.loadFS(config.Path)
	if err != nil {
		return nil, fmt.Errorf("could not load FS repository: %w", err)
	}

	return r, nil
}

// loadFS will load all the resource and groups from the root FS path.
// the loaded resources will be loaded into a memory repository.
func (r *Repository) loadFS(rootPath string) error {
	// Create a regex that will strip the group IDs from the path by extracting the manifests root path
	// from the paths (scapes the common path characters like `.` or `/` for the regex).
	scapedRootPath := strings.ReplaceAll(rootPath, ".", `\.`)
	scapedRootPath = strings.ReplaceAll(scapedRootPath, "/", `\/`)
	groupIDRegexReplace, err := regexp.Compile(fmt.Sprintf(`%s\/?`, scapedRootPath))
	if err != nil {
		return fmt.Errorf("could not compile groupID regex: %w", err)
	}

	// Walk doesn't apply concurrency, its safe to mutate these variables in
	// this context from the walkFn context.
	groups := map[string]model.Group{}
	resources := map[string]model.Resource{}

	err = r.fsManager.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		logger := r.logger.WithValues(log.Kv{"path": path})
		if err != nil {
			logger.Warningf("could not access a path: %s", err)
			return err
		}

		// Directories and non YAML files don't need to be handled.
		extension := strings.ToLower(filepath.Ext(path))
		if info.IsDir() || (extension != ".yml" && extension != ".yaml") {
			return nil
		}

		// Check if we need to ignore, using absolute path.
		absPath, err := r.fsManager.Abs(path)
		if err != nil {
			return err
		}
		if r.shouldIgnore(absPath) {
			logger.Debugf("ignoring file: %q", absPath)
			return nil
		}

		// Get group information.
		groupPath := filepath.Dir(path)
		groupID := groupIDRegexReplace.ReplaceAllString(filepath.Dir(path), "")
		if groupID == "" {
			groupID = r.rootGroupID
		}

		// Read file and load kubernetes objects.
		objs, err := r.loadK8sObjects(path)
		if err != nil {
			return err
		}
		for _, obj := range objs {
			resource, err := r.modelFactory.NewResource(obj, groupID, path)
			if err != nil {
				return fmt.Errorf("could not create model resource: %w", err)
			}

			// Check if we already have this resource.
			stored, ok := resources[resource.ID]
			if ok {
				return fmt.Errorf("%w: resource collision with %s in %q and %q", internalerrors.ErrNotValid, resource.ID, stored.ManifestPath, path)
			}

			// Store the resource.
			resources[resource.ID] = *resource
		}

		// Check if we already have the group and check if two different groups have the same name.
		groupConfig := r.appConfig.Groups[groupID]
		groups[groupID] = r.modelFactory.NewGroup(groupID, groupPath, groupConfig)

		return nil
	})
	if err != nil {
		return fmt.Errorf("could not load fs manifests: %w", err)
	}

	r.resourceMemoryRepo = storagememory.NewResourceRepository(resources)
	r.groupMemoryRepo = storagememory.NewGroupRepository(groups)

	return nil
}

func (r *Repository) shouldIgnore(path string) bool {
	for _, rgx := range r.excludeRegex {
		if rgx.MatchString(path) {
			return true
		}
	}

	for _, rgx := range r.includeRegex {
		if rgx.MatchString(path) {
			return false
		}
	}

	return r.defaultIgnore
}

func (r *Repository) loadK8sObjects(path string) ([]model.K8sObject, error) {
	fileData, err := r.fsManager.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read %q file: %w", path, err)
	}
	objs, err := r.k8sDecoder.DecodeObjects(context.Background(), fileData)
	if err != nil {
		return nil, fmt.Errorf("could not load kubernetes objects in %s: %w", path, err)
	}

	return objs, nil
}

// GetResource satisfies storage.ResourceRepository interface.
func (r *Repository) GetResource(ctx context.Context, id string) (*model.Resource, error) {
	return r.resourceMemoryRepo.GetResource(ctx, id)
}

// ListResources satisfies storage.ResourceRepository interface.
func (r *Repository) ListResources(ctx context.Context, opts storage.ResourceListOpts) (*storage.ResourceList, error) {
	return r.resourceMemoryRepo.ListResources(ctx, opts)
}

// GetGroup satisfies storage.GroupRepository interface.
func (r *Repository) GetGroup(ctx context.Context, id string) (*model.Group, error) {
	return r.groupMemoryRepo.GetGroup(ctx, id)
}

// ListGroups satisfies storage.GroupRepository interface.
func (r *Repository) ListGroups(ctx context.Context, opts storage.GroupListOpts) (*storage.GroupList, error) {
	return r.groupMemoryRepo.ListGroups(ctx, opts)
}
