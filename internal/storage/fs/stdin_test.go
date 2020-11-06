package fs_test

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/storage"
	"github.com/slok/kahoy/internal/storage/fs"
	"github.com/slok/kahoy/internal/storage/fs/fsmock"
)

func TestIOReaderRepositoryLoad(t *testing.T) {
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	tests := map[string]struct {
		cfg          fs.IOReaderRepositoryConfig
		mock         func(mkd *fsmock.K8sObjectDecoder)
		expResources []model.Resource
		expGroups    []model.Group
		expErr       bool
	}{
		"If no reader given, it should fail.": {
			cfg: fs.IOReaderRepositoryConfig{
				AppConfig: &model.AppConfig{},
			},
			mock:         func(mkd *fsmock.K8sObjectDecoder) {},
			expResources: []model.Resource{},
			expGroups:    []model.Group{},
			expErr:       true,
		},

		"Data with resources should return the resource and groups.": {
			cfg: fs.IOReaderRepositoryConfig{
				AppConfig: &model.AppConfig{},
				Reader:    strings.NewReader(`somedata`),
			},
			mock: func(mkd *fsmock.K8sObjectDecoder) {
				objs := []model.K8sObject{
					newConfigmap("test-ns", "test-name"),
					newConfigmap("test-ns2", "test-name2"),
				}
				mkd.On("DecodeObjects", mock.Anything, []byte("somedata")).Once().Return(objs, nil)
			},
			expResources: []model.Resource{
				{
					ID:           "core/v1/ConfigMap/test-ns/test-name",
					GroupID:      "root",
					ManifestPath: "stdin",
					K8sObject:    newConfigmap("test-ns", "test-name"),
				},
				{
					ID:           "core/v1/ConfigMap/test-ns2/test-name2",
					GroupID:      "root",
					ManifestPath: "stdin",
					K8sObject:    newConfigmap("test-ns2", "test-name2"),
				},
			},
			expGroups: []model.Group{
				{ID: "root", Path: "", Priority: 1000},
			},
			expErr: false,
		},

		"Repeated resources on data, should fail.": {
			cfg: fs.IOReaderRepositoryConfig{
				AppConfig: &model.AppConfig{},
				Reader:    strings.NewReader(`somedata`),
			},
			mock: func(mkd *fsmock.K8sObjectDecoder) {
				objs := []model.K8sObject{
					newConfigmap("test-ns", "test-name"),
					newConfigmap("test-ns", "test-name"),
				}
				mkd.On("DecodeObjects", mock.Anything, mock.Anything).Once().Return(objs, nil)
			},
			expErr: true,
		},

		"Having errors when loading resources should fail.": {
			cfg: fs.IOReaderRepositoryConfig{
				AppConfig: &model.AppConfig{},
				Reader:    strings.NewReader(`somedata`),
			},
			mock: func(mkd *fsmock.K8sObjectDecoder) {
				mkd.On("DecodeObjects", mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("whatever"))
			},
			expErr: true,
		},

		"Context done should end loading process.": {
			cfg: fs.IOReaderRepositoryConfig{
				Ctx:       cancelledCtx,
				AppConfig: &model.AppConfig{},
				Reader:    strings.NewReader(`somedata`),
			},
			mock:   func(mkd *fsmock.K8sObjectDecoder) {},
			expErr: true,
		},

		"A timeout while loading should error.": {
			cfg: fs.IOReaderRepositoryConfig{
				LoadTimeout: 1,
				AppConfig:   &model.AppConfig{},
				Reader:      strings.NewReader(`somedata`),
			},
			mock:   func(mkd *fsmock.K8sObjectDecoder) {},
			expErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Mocks.
			mkd := &fsmock.K8sObjectDecoder{}
			test.mock(mkd)

			// Load and check errors.
			test.cfg.KubernetesDecoder = mkd
			test.cfg.ModelFactory = newModelResourceAndGroupFactory()
			repo, err := fs.NewIOReaderRepository(test.cfg)

			if test.expErr {
				assert.Error(err)
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
