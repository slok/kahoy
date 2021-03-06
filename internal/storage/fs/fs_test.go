package fs_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/model/modelmock"
	"github.com/slok/kahoy/internal/storage"
	"github.com/slok/kahoy/internal/storage/fs"
	"github.com/slok/kahoy/internal/storage/fs/fsmock"
)

// Helper alias for verbosity of unstructured internal maps.
type tm = map[string]interface{}

// testInfoFile is an easy way to create fake test info files.
type testInfoFile struct {
	name  string
	isDir bool
}

func (t testInfoFile) Name() string       { return t.name }
func (t testInfoFile) Size() int64        { return 0 }
func (t testInfoFile) Mode() os.FileMode  { return 0 }
func (t testInfoFile) ModTime() time.Time { return time.Now() }
func (t testInfoFile) IsDir() bool        { return t.isDir }
func (t testInfoFile) Sys() interface{}   { return nil }

var _ os.FileInfo = &testInfoFile{}

func newModelResourceAndGroupFactory() *model.ResourceAndGroupFactory {
	mk := &modelmock.KubernetesDiscoveryClient{}
	mk.On("GetServerGroupsAndResources", mock.Anything).Return(nil, []*metav1.APIResourceList{
		{
			GroupVersion: "v1",
			APIResources: []metav1.APIResource{
				{Kind: "ConfigMap", Namespaced: true},
			},
		}}, nil)

	f, _ := model.NewResourceAndGroupFactory(mk, log.Noop)
	return f
}

func newConfigmap(ns, name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: tm{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": tm{
				"name":      name,
				"namespace": ns,
			},
		},
	}
}

// TestRepositoryLoadFS will test that an manifest FS based repository
// is loaded correctly, it mocks all the FS so we are testing the internal
// implementation, but this is ok because is a low level component.
func TestRepositoryLoadFS(t *testing.T) {
	var errWanted = errors.New("wanted error")

	tests := map[string]struct {
		cfg          fs.RepositoryConfig
		mock         func(mfsm *fsmock.FileSystemManager, mkd *fsmock.K8sObjectDecoder)
		expResources []model.Resource
		expGroups    []model.Group
		expErr       error
	}{
		"Having an error while loading the FS, should return the error.": {
			cfg: fs.RepositoryConfig{
				Path:      "/tmp/test",
				AppConfig: &model.AppConfig{},
			},
			mock: func(mfsm *fsmock.FileSystemManager, mkd *fsmock.K8sObjectDecoder) {
				mfsm.On("Walk", mock.Anything, mock.Anything).Once().Return(errWanted)
			},
			expErr: errWanted,
		},

		"Having no files should not load resources.": {
			cfg: fs.RepositoryConfig{
				AppConfig: &model.AppConfig{},
				Path:      "/tmp/test",
			},
			mock: func(mfsm *fsmock.FileSystemManager, mkd *fsmock.K8sObjectDecoder) {
				mfsm.On("Walk", mock.Anything, mock.Anything).Once().Return(nil)
			},
			expResources: []model.Resource{},
			expGroups:    []model.Group{},
		},

		"Having a file with one resource present resource one single resource should be returned.": {
			cfg: fs.RepositoryConfig{
				AppConfig: &model.AppConfig{},
				Path:      "/tmp/test",
			},
			mock: func(mfsm *fsmock.FileSystemManager, mkd *fsmock.K8sObjectDecoder) {
				// Mock 1 file.
				f1Path := "/tmp/test/group1/test-1.yaml"
				f1 := testInfoFile{name: "test-1.yaml", isDir: false}
				mfsm.On("Abs", "/tmp/test/group1/test-1.yaml").Once().Return("/tmp/test/group1/test-1.yaml", nil)
				mfsm.On("ReadFile", "/tmp/test/group1/test-1.yaml").Once().Return([]byte("f1"), nil)
				objs := []model.K8sObject{newConfigmap("test-ns", "test-name")}
				mkd.On("DecodeObjects", mock.Anything, []byte("f1")).Once().Return(objs, nil)

				// Mock all fs walks that will trigger the other mocks.
				mfsm.On("Walk", "/tmp/test", mock.Anything).Once().Return(nil).Run(func(args mock.Arguments) {
					walkfn := args[1].(filepath.WalkFunc)
					_ = walkfn(f1Path, f1, nil)
				})
			},
			expResources: []model.Resource{
				{
					ID:           "core/v1/ConfigMap/test-ns/test-name",
					GroupID:      "group1",
					ManifestPath: "/tmp/test/group1/test-1.yaml",
					K8sObject:    newConfigmap("test-ns", "test-name"),
				},
			},
			expGroups: []model.Group{
				{ID: "group1", Path: "/tmp/test/group1", Priority: 1000},
			},
		},

		"Having a file with multiple resources present resource one single resource should be load all resources with a single group.": {
			cfg: fs.RepositoryConfig{
				AppConfig: &model.AppConfig{},
				Path:      "/tmp/test",
			},
			mock: func(mfsm *fsmock.FileSystemManager, mkd *fsmock.K8sObjectDecoder) {
				// Mock 1 file.
				f1Path := "/tmp/test/group1/test-1.yaml"
				f1 := testInfoFile{name: "test-1.yaml", isDir: false}
				mfsm.On("Abs", "/tmp/test/group1/test-1.yaml").Once().Return("/tmp/test/group1/test-1.yaml", nil)
				mfsm.On("ReadFile", "/tmp/test/group1/test-1.yaml").Once().Return([]byte("f1"), nil)
				objs := []model.K8sObject{
					newConfigmap("test-ns", "test-name"),
					newConfigmap("test-ns2", "test-name2"),
				}
				mkd.On("DecodeObjects", mock.Anything, []byte("f1")).Once().Return(objs, nil)

				// Mock all fs walks that will trigger the other mocks.
				mfsm.On("Walk", "/tmp/test", mock.Anything).Once().Return(nil).Run(func(args mock.Arguments) {
					walkfn := args[1].(filepath.WalkFunc)
					_ = walkfn(f1Path, f1, nil)
				})
			},
			expResources: []model.Resource{
				{
					ID:           "core/v1/ConfigMap/test-ns/test-name",
					GroupID:      "group1",
					ManifestPath: "/tmp/test/group1/test-1.yaml",
					K8sObject:    newConfigmap("test-ns", "test-name"),
				},
				{
					ID:           "core/v1/ConfigMap/test-ns2/test-name2",
					GroupID:      "group1",
					ManifestPath: "/tmp/test/group1/test-1.yaml",
					K8sObject:    newConfigmap("test-ns2", "test-name2"),
				},
			},
			expGroups: []model.Group{
				{ID: "group1", Path: "/tmp/test/group1", Priority: 1000},
			},
		},

		"Having multiple files in same group should load all resources with the single group.": {
			cfg: fs.RepositoryConfig{
				AppConfig: &model.AppConfig{},
				Path:      "/tmp/test",
			},
			mock: func(mfsm *fsmock.FileSystemManager, mkd *fsmock.K8sObjectDecoder) {
				// Mock 2 files.
				f1Path := "/tmp/test/group1/test-1.yaml"
				f1 := testInfoFile{name: "test-1.yaml", isDir: false}
				mfsm.On("Abs", "/tmp/test/group1/test-1.yaml").Once().Return("/tmp/test/group1/test-1.yaml", nil)
				mfsm.On("ReadFile", "/tmp/test/group1/test-1.yaml").Once().Return([]byte("f1"), nil)
				objs := []model.K8sObject{
					newConfigmap("test-ns", "test-name"),
				}
				mkd.On("DecodeObjects", mock.Anything, []byte("f1")).Once().Return(objs, nil)

				f2Path := "/tmp/test/group1/test-2.yaml"
				f2 := testInfoFile{name: "test-2.yaml", isDir: false}
				mfsm.On("Abs", "/tmp/test/group1/test-2.yaml").Once().Return("/tmp/test/group1/test-2.yaml", nil)
				mfsm.On("ReadFile", "/tmp/test/group1/test-2.yaml").Once().Return([]byte("f2"), nil)
				objs2 := []model.K8sObject{
					newConfigmap("test-ns2", "test-name2"),
				}
				mkd.On("DecodeObjects", mock.Anything, []byte("f2")).Once().Return(objs2, nil)

				// Mock all fs walks that will trigger the other mocks.
				mfsm.On("Walk", "/tmp/test", mock.Anything).Once().Return(nil).Run(func(args mock.Arguments) {
					walkfn := args[1].(filepath.WalkFunc)
					_ = walkfn(f1Path, f1, nil)
					_ = walkfn(f2Path, f2, nil)
				})
			},
			expResources: []model.Resource{
				{
					ID:           "core/v1/ConfigMap/test-ns/test-name",
					GroupID:      "group1",
					ManifestPath: "/tmp/test/group1/test-1.yaml",
					K8sObject:    newConfigmap("test-ns", "test-name"),
				},
				{
					ID:           "core/v1/ConfigMap/test-ns2/test-name2",
					GroupID:      "group1",
					ManifestPath: "/tmp/test/group1/test-2.yaml",
					K8sObject:    newConfigmap("test-ns2", "test-name2"),
				},
			},
			expGroups: []model.Group{
				{ID: "group1", Path: "/tmp/test/group1", Priority: 1000},
			},
		},

		"Having multiple files in different groups (and subdirectories) should be load all resources with multiple group.": {
			cfg: fs.RepositoryConfig{
				AppConfig: &model.AppConfig{},
				Path:      "/tmp/test",
			},
			mock: func(mfsm *fsmock.FileSystemManager, mkd *fsmock.K8sObjectDecoder) {
				// Mock files.
				f1Path := "/tmp/test/group1/test-1.yaml"
				f1 := testInfoFile{name: "test-1.yaml", isDir: false}
				mfsm.On("Abs", "/tmp/test/group1/test-1.yaml").Once().Return("/tmp/test/group1/test-1.yaml", nil)
				mfsm.On("ReadFile", "/tmp/test/group1/test-1.yaml").Once().Return([]byte("f1"), nil)
				objs := []model.K8sObject{
					newConfigmap("test-ns", "test-name"),
				}
				mkd.On("DecodeObjects", mock.Anything, []byte("f1")).Once().Return(objs, nil)

				f2Path := "/tmp/test/group2/test-2.yaml"
				f2 := testInfoFile{name: "test-2.yaml", isDir: false}
				mfsm.On("Abs", "/tmp/test/group2/test-2.yaml").Once().Return("/tmp/test/group2/test-2.yaml", nil)
				mfsm.On("ReadFile", "/tmp/test/group2/test-2.yaml").Once().Return([]byte("f2"), nil)
				objs2 := []model.K8sObject{
					newConfigmap("test-ns2", "test-name2"),
				}
				mkd.On("DecodeObjects", mock.Anything, []byte("f2")).Once().Return(objs2, nil)

				f3Path := "/tmp/test/group2/subgroup3/test-3.yaml"
				f3 := testInfoFile{name: "test-3.yaml", isDir: false}
				mfsm.On("Abs", "/tmp/test/group2/subgroup3/test-3.yaml").Once().Return("/tmp/test/group2/subgroup3/test-3.yaml", nil)
				mfsm.On("ReadFile", "/tmp/test/group2/subgroup3/test-3.yaml").Once().Return([]byte("f3"), nil)
				objs3 := []model.K8sObject{
					newConfigmap("test-ns3", "test-name3"),
				}
				mkd.On("DecodeObjects", mock.Anything, []byte("f3")).Once().Return(objs3, nil)

				// Mock all fs walks that will trigger the other mocks.
				mfsm.On("Walk", "/tmp/test", mock.Anything).Once().Return(nil).Run(func(args mock.Arguments) {
					walkfn := args[1].(filepath.WalkFunc)
					_ = walkfn(f1Path, f1, nil)
					_ = walkfn(f2Path, f2, nil)
					_ = walkfn(f3Path, f3, nil)
				})
			},
			expResources: []model.Resource{
				{
					ID:           "core/v1/ConfigMap/test-ns/test-name",
					GroupID:      "group1",
					ManifestPath: "/tmp/test/group1/test-1.yaml",
					K8sObject:    newConfigmap("test-ns", "test-name"),
				},
				{
					ID:           "core/v1/ConfigMap/test-ns2/test-name2",
					GroupID:      "group2",
					ManifestPath: "/tmp/test/group2/test-2.yaml",
					K8sObject:    newConfigmap("test-ns2", "test-name2"),
				},
				{
					ID:           "core/v1/ConfigMap/test-ns3/test-name3",
					GroupID:      "group2/subgroup3",
					ManifestPath: "/tmp/test/group2/subgroup3/test-3.yaml",
					K8sObject:    newConfigmap("test-ns3", "test-name3"),
				},
			},
			expGroups: []model.Group{
				{ID: "group1", Path: "/tmp/test/group1", Priority: 1000},
				{ID: "group2", Path: "/tmp/test/group2", Priority: 1000},
				{ID: "group2/subgroup3", Path: "/tmp/test/group2/subgroup3", Priority: 1000},
			},
		},

		"Having files in root path, should be on the default root group id.": {
			cfg: fs.RepositoryConfig{
				AppConfig: &model.AppConfig{},
				Path:      "/tmp/test",
			},
			mock: func(mfsm *fsmock.FileSystemManager, mkd *fsmock.K8sObjectDecoder) {
				// Mock 1 file.
				f1Path := "/tmp/test/test-1.yaml"
				f1 := testInfoFile{name: "test-1.yaml", isDir: false}
				mfsm.On("Abs", "/tmp/test/test-1.yaml").Once().Return("/tmp/test/test-1.yaml", nil)
				mfsm.On("ReadFile", "/tmp/test/test-1.yaml").Once().Return([]byte("f1"), nil)
				objs := []model.K8sObject{newConfigmap("test-ns", "test-name")}
				mkd.On("DecodeObjects", mock.Anything, []byte("f1")).Once().Return(objs, nil)

				// Mock all fs walks that will trigger the other mocks.
				mfsm.On("Walk", "/tmp/test", mock.Anything).Once().Return(nil).Run(func(args mock.Arguments) {
					walkfn := args[1].(filepath.WalkFunc)
					_ = walkfn(f1Path, f1, nil)
				})
			},
			expResources: []model.Resource{
				{
					ID:           "core/v1/ConfigMap/test-ns/test-name",
					GroupID:      "root",
					ManifestPath: "/tmp/test/test-1.yaml",
					K8sObject:    newConfigmap("test-ns", "test-name"),
				},
			},
			expGroups: []model.Group{
				{ID: "root", Path: "/tmp/test", Priority: 1000},
			},
		},

		"Having files in root path (custom), should be on the default root path.": {
			cfg: fs.RepositoryConfig{
				AppConfig:   &model.AppConfig{},
				Path:        "/tmp/test",
				RootGroupID: "whatever",
			},
			mock: func(mfsm *fsmock.FileSystemManager, mkd *fsmock.K8sObjectDecoder) {
				// Mock 1 file.
				f1Path := "/tmp/test/test-1.yaml"
				f1 := testInfoFile{name: "test-1.yaml", isDir: false}
				mfsm.On("Abs", "/tmp/test/test-1.yaml").Once().Return("/tmp/test/test-1.yaml", nil)
				mfsm.On("ReadFile", "/tmp/test/test-1.yaml").Once().Return([]byte("f1"), nil)
				objs := []model.K8sObject{newConfigmap("test-ns", "test-name")}
				mkd.On("DecodeObjects", mock.Anything, []byte("f1")).Once().Return(objs, nil)

				// Mock all fs walks that will trigger the other mocks.
				mfsm.On("Walk", "/tmp/test", mock.Anything).Once().Return(nil).Run(func(args mock.Arguments) {
					walkfn := args[1].(filepath.WalkFunc)
					_ = walkfn(f1Path, f1, nil)
				})
			},
			expResources: []model.Resource{
				{
					ID:           "core/v1/ConfigMap/test-ns/test-name",
					GroupID:      "whatever",
					ManifestPath: "/tmp/test/test-1.yaml",
					K8sObject:    newConfigmap("test-ns", "test-name"),
				},
			},
			expGroups: []model.Group{
				{ID: "whatever", Path: "/tmp/test", Priority: 1000},
			},
		},

		"Having a non yaml file, should be ignored.": {
			cfg: fs.RepositoryConfig{
				AppConfig: &model.AppConfig{},
				Path:      "/tmp/test",
			},
			mock: func(mfsm *fsmock.FileSystemManager, mkd *fsmock.K8sObjectDecoder) {
				// Mock 1 file.
				f1Path := "/tmp/test/group1/test-1.json"
				f1 := testInfoFile{name: "test-1.json", isDir: false}

				// Mock all fs walks that will trigger the other mocks.
				mfsm.On("Walk", "/tmp/test", mock.Anything).Once().Return(nil).Run(func(args mock.Arguments) {
					walkfn := args[1].(filepath.WalkFunc)
					_ = walkfn(f1Path, f1, nil)
				})
			},
			expResources: []model.Resource{},
			expGroups:    []model.Group{},
		},

		"Directories and excluded regex should be ignored.": {
			cfg: fs.RepositoryConfig{
				AppConfig:    &model.AppConfig{},
				Path:         "/tmp/test",
				ExcludeRegex: []string{".*/group2", ""},
			},
			mock: func(mfsm *fsmock.FileSystemManager, mkd *fsmock.K8sObjectDecoder) {
				// Mock dirs.
				f1Path := "/tmp/test/group1"
				f1 := testInfoFile{name: "group1", isDir: true}
				f2Path := "/tmp/test/group2/test2.yaml"
				f2 := testInfoFile{name: "group2"}
				f3Path := "/tmp/test/group3/test3.yaml"
				f3 := testInfoFile{name: "test3.yaml"}

				// The ignored file.
				mfsm.On("Abs", "/tmp/test/group2/test2.yaml").Once().Return("/tmp/test/group2/test2.yaml", nil)

				// The file that is included.
				mfsm.On("ReadFile", "/tmp/test/group3/test3.yaml").Once().Return([]byte("f3"), nil)
				mfsm.On("Abs", "/tmp/test/group3/test3.yaml").Once().Return("/tmp/test/group3/test3.yaml", nil)
				objs := []model.K8sObject{
					newConfigmap("test-ns", "test-name3"),
				}
				mkd.On("DecodeObjects", mock.Anything, []byte("f3")).Once().Return(objs, nil)

				// Mock all fs walks that will trigger the other mocks.
				mfsm.On("Walk", "/tmp/test", mock.Anything).Once().Return(nil).Run(func(args mock.Arguments) {
					walkfn := args[1].(filepath.WalkFunc)
					_ = walkfn(f1Path, f1, nil)
					_ = walkfn(f2Path, f2, nil)
					_ = walkfn(f3Path, f3, nil)
				})
			},
			expResources: []model.Resource{
				{
					ID:           "core/v1/ConfigMap/test-ns/test-name3",
					GroupID:      "group3",
					ManifestPath: "/tmp/test/group3/test3.yaml",
					K8sObject:    newConfigmap("test-ns", "test-name3"),
				},
			},
			expGroups: []model.Group{
				{ID: "group3", Path: "/tmp/test/group3", Priority: 1000},
			},
		},

		"Included should be included and ignore others.": {
			cfg: fs.RepositoryConfig{
				AppConfig:    &model.AppConfig{},
				Path:         "/tmp/test",
				IncludeRegex: []string{".*/group2"},
			},
			mock: func(mfsm *fsmock.FileSystemManager, mkd *fsmock.K8sObjectDecoder) {
				// Mock 3 files.
				f1Path := "/tmp/test/group1/test-1.yaml"
				mfsm.On("Abs", "/tmp/test/group1/test-1.yaml").Once().Return("/tmp/test/group1/test-1.yaml", nil)
				f1 := testInfoFile{name: "test1", isDir: false}
				f2Path := "/tmp/test/group2/test-1.yaml"
				mfsm.On("Abs", "/tmp/test/group2/test-1.yaml").Once().Return("/tmp/test/group2/test-1.yaml", nil)
				f2 := testInfoFile{name: "test2", isDir: false}
				f3Path := "/tmp/test/group3/test-1.yaml"
				mfsm.On("Abs", "/tmp/test/group3/test-1.yaml").Once().Return("/tmp/test/group3/test-1.yaml", nil)
				f3 := testInfoFile{name: "test3", isDir: false}

				// The file that is included.
				mfsm.On("ReadFile", "/tmp/test/group2/test-1.yaml").Once().Return([]byte("f1"), nil)
				objs := []model.K8sObject{
					newConfigmap("test-ns", "test-name"),
				}
				mkd.On("DecodeObjects", mock.Anything, []byte("f1")).Once().Return(objs, nil)

				// Mock all fs walks that will trigger the other mocks.
				mfsm.On("Walk", "/tmp/test", mock.Anything).Once().Return(nil).Run(func(args mock.Arguments) {
					walkfn := args[1].(filepath.WalkFunc)
					_ = walkfn(f1Path, f1, nil)
					_ = walkfn(f2Path, f2, nil)
					_ = walkfn(f3Path, f3, nil)
				})
			},
			expResources: []model.Resource{
				{
					ID:           "core/v1/ConfigMap/test-ns/test-name",
					GroupID:      "group2",
					ManifestPath: "/tmp/test/group2/test-1.yaml",
					K8sObject:    newConfigmap("test-ns", "test-name"),
				},
			},
			expGroups: []model.Group{
				{ID: "group2", Path: "/tmp/test/group2", Priority: 1000},
			},
		},

		"Excludes should have priority over includes.": {
			cfg: fs.RepositoryConfig{
				AppConfig:    &model.AppConfig{},
				Path:         "/tmp/test",
				ExcludeRegex: []string{".*/group2"},
				IncludeRegex: []string{".*/group2"},
			},
			mock: func(mfsm *fsmock.FileSystemManager, mkd *fsmock.K8sObjectDecoder) {
				// Mock 3 files.
				f1Path := "/tmp/test/group1/test-1.yaml"
				mfsm.On("Abs", "/tmp/test/group1/test-1.yaml").Once().Return("/tmp/test/group1/test-1.yaml", nil)
				f1 := testInfoFile{name: "test1", isDir: false}
				f2Path := "/tmp/test/group2/test-1.yaml"
				mfsm.On("Abs", "/tmp/test/group2/test-1.yaml").Once().Return("/tmp/test/group2/test-1.yaml", nil)
				f2 := testInfoFile{name: "test2", isDir: false}
				f3Path := "/tmp/test/group3/test-1.yaml"
				mfsm.On("Abs", "/tmp/test/group3/test-1.yaml").Once().Return("/tmp/test/group3/test-1.yaml", nil)
				f3 := testInfoFile{name: "test3", isDir: false}

				// Mock all fs walks that will trigger the other mocks.
				mfsm.On("Walk", "/tmp/test", mock.Anything).Once().Return(nil).Run(func(args mock.Arguments) {
					walkfn := args[1].(filepath.WalkFunc)
					_ = walkfn(f1Path, f1, nil)
					_ = walkfn(f2Path, f2, nil)
					_ = walkfn(f3Path, f3, nil)
				})
			},
			expResources: []model.Resource{},
			expGroups:    []model.Group{},
		},

		"An error on walk should stop the walk on the FS.": {
			cfg: fs.RepositoryConfig{
				AppConfig: &model.AppConfig{},
				Path:      "/tmp/test",
			},
			mock: func(mfsm *fsmock.FileSystemManager, mkd *fsmock.K8sObjectDecoder) {
				// Mock 1 file.
				f1Path := "/tmp/test/group1/test-1.yaml"
				f1 := testInfoFile{name: "test-1.yaml", isDir: false}

				// Mock all fs walks that will trigger the other mocks.
				mfsm.On("Walk", "/tmp/test", mock.Anything).Once().Return(nil).Run(func(args mock.Arguments) {
					walkfn := args[1].(filepath.WalkFunc)
					_ = walkfn(f1Path, f1, errWanted)
				})
			},
			expResources: []model.Resource{},
			expGroups:    []model.Group{},
		},

		"Having ignores on files, it should ignore matched files.": {
			cfg: fs.RepositoryConfig{
				AppConfig:    &model.AppConfig{},
				Path:         "/tmp/test",
				ExcludeRegex: []string{".*test-2.*"},
			},
			mock: func(mfsm *fsmock.FileSystemManager, mkd *fsmock.K8sObjectDecoder) {
				// Mock 1 file.
				f1Path := "/tmp/test/group1/test-1.yaml"
				f1 := testInfoFile{name: "test-1.yaml", isDir: false}
				mfsm.On("Abs", "/tmp/test/group1/test-1.yaml").Once().Return("/tmp/test/group1/test-1.yaml", nil)
				mfsm.On("ReadFile", "/tmp/test/group1/test-1.yaml").Once().Return([]byte("f1"), nil)
				objs := []model.K8sObject{
					newConfigmap("test-ns", "test-name"),
				}
				mkd.On("DecodeObjects", mock.Anything, []byte("f1")).Once().Return(objs, nil)

				// Ignored file.
				f2Path := "/tmp/test/group1/test-2.yaml"
				mfsm.On("Abs", "/tmp/test/group1/test-2.yaml").Once().Return("/tmp/test/group1/test-2.yaml", nil)
				f2 := testInfoFile{name: "test-2.yaml", isDir: false}

				// Mock all fs walks that will trigger the other mocks.
				mfsm.On("Walk", "/tmp/test", mock.Anything).Once().Return(nil).Run(func(args mock.Arguments) {
					walkfn := args[1].(filepath.WalkFunc)
					_ = walkfn(f1Path, f1, nil)
					_ = walkfn(f2Path, f2, nil)
				})
			},
			expResources: []model.Resource{
				{
					ID:           "core/v1/ConfigMap/test-ns/test-name",
					GroupID:      "group1",
					ManifestPath: "/tmp/test/group1/test-1.yaml",
					K8sObject:    newConfigmap("test-ns", "test-name"),
				},
			},
			expGroups: []model.Group{
				{ID: "group1", Path: "/tmp/test/group1", Priority: 1000},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Mocks.
			mfsm := &fsmock.FileSystemManager{}
			mkd := &fsmock.K8sObjectDecoder{}
			test.mock(mfsm, mkd)

			test.cfg.FSManager = mfsm
			test.cfg.KubernetesDecoder = mkd

			// Load and check errors.
			test.cfg.ModelFactory = newModelResourceAndGroupFactory()
			repo, err := fs.NewRepository(test.cfg)
			if test.expErr != nil {
				assert.True(errors.Is(err, test.expErr))
			} else if assert.NoError(err) {
				// Check loaded resources.
				gotResources, err := repo.ListResources(context.TODO(), storage.ResourceListOpts{})
				require.NoError(err)
				sort.SliceStable(test.expResources, func(i, j int) bool { return test.expResources[i].ID < test.expResources[j].ID })
				sort.SliceStable(gotResources.Items, func(i, j int) bool { return gotResources.Items[i].ID < gotResources.Items[j].ID })
				assert.Equal(test.expResources, gotResources.Items)

				// Check loaded groups.
				gotGroups, err := repo.ListGroups(context.TODO(), storage.GroupListOpts{})
				require.NoError(err)
				sort.SliceStable(test.expGroups, func(i, j int) bool { return test.expGroups[i].ID < test.expGroups[j].ID })
				sort.SliceStable(gotGroups.Items, func(i, j int) bool { return gotGroups.Items[i].ID < gotGroups.Items[j].ID })
				assert.Equal(test.expGroups, gotGroups.Items)
			}
		})
	}
}
