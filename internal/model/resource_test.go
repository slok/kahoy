package model_test

import (
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
)

func TestResourceAndGroupFactoryNewResource(t *testing.T) {
	// Helper alias for verbosity of unstructured internal maps.
	type tm = map[string]interface{}

	testAPIResourceList := []*metav1.APIResourceList{
		{
			GroupVersion: "networking.k8s.io/v1beta1",
			APIResources: []metav1.APIResource{
				{Kind: "Ingress", Namespaced: true},
			},
		},
		{
			GroupVersion: "v1",
			APIResources: []metav1.APIResource{
				{Kind: "Pod", Namespaced: true},
			},
		},
		{
			GroupVersion: "rbac.authorization.k8s.io/v1",
			APIResources: []metav1.APIResource{
				{Kind: "ClusterRoleBinding", Namespaced: false},
			},
		},
	}

	tests := map[string]struct {
		groupID      string
		manifestPath string
		obj          model.K8sObject
		mock         func(m *modelmock.KubernetesDiscoveryClient)
		expResource  model.Resource
		expErr       bool
	}{
		"A type that is unknown by the apiserver should fail.": {
			obj: &unstructured.Unstructured{
				Object: tm{
					"apiVersion": "v1",
					"kind":       "Service",
					"metadata": tm{
						"name":      "test-name",
						"namespace": "test-ns",
					},
				},
			},
			groupID:      "test-group",
			manifestPath: "/test",
			mock: func(m *modelmock.KubernetesDiscoveryClient) {
				m.On("GetServerGroupsAndResources", mock.Anything).Once().Return(nil, testAPIResourceList, nil)
			},
			expErr: true,
		},

		"A namespaced known type should return correctly the resource.": {
			obj: &unstructured.Unstructured{
				Object: tm{
					"apiVersion": "networking.k8s.io/v1beta1",
					"kind":       "Ingress",
					"metadata": tm{
						"name":      "test-name",
						"namespace": "test-ns",
					},
				},
			},
			groupID:      "test-group",
			manifestPath: "/test",
			mock: func(m *modelmock.KubernetesDiscoveryClient) {
				m.On("GetServerGroupsAndResources", mock.Anything).Once().Return(nil, testAPIResourceList, nil)
			},
			expResource: model.Resource{
				ID:           "networking.k8s.io/v1beta1/Ingress/test-ns/test-name",
				GroupID:      "test-group",
				ManifestPath: "/test",
				K8sObject: &unstructured.Unstructured{
					Object: tm{
						"apiVersion": "networking.k8s.io/v1beta1",
						"kind":       "Ingress",
						"metadata": tm{
							"name":      "test-name",
							"namespace": "test-ns",
						},
					},
				},
			},
		},

		"A namespaced known type without namespace set, should return correctly the resource.": {
			obj: &unstructured.Unstructured{
				Object: tm{
					"apiVersion": "networking.k8s.io/v1beta1",
					"kind":       "Ingress",
					"metadata": tm{
						"name": "test-name",
					},
				},
			},
			groupID:      "test-group",
			manifestPath: "/test",
			mock: func(m *modelmock.KubernetesDiscoveryClient) {
				m.On("GetServerGroupsAndResources", mock.Anything).Once().Return(nil, testAPIResourceList, nil)
			},
			expResource: model.Resource{
				ID:           "networking.k8s.io/v1beta1/Ingress/default/test-name",
				GroupID:      "test-group",
				ManifestPath: "/test",
				K8sObject: &unstructured.Unstructured{
					Object: tm{
						"apiVersion": "networking.k8s.io/v1beta1",
						"kind":       "Ingress",
						"metadata": tm{
							"name": "test-name",
						},
					},
				},
			},
		},

		"A core known type, should return correctly the resource.": {
			obj: &unstructured.Unstructured{
				Object: tm{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": tm{
						"name":      "test-name",
						"namespace": "test-ns",
					},
				},
			},
			groupID:      "test-group",
			manifestPath: "/test",
			mock: func(m *modelmock.KubernetesDiscoveryClient) {
				m.On("GetServerGroupsAndResources", mock.Anything).Once().Return(nil, testAPIResourceList, nil)
			},
			expResource: model.Resource{
				ID:           "core/v1/Pod/test-ns/test-name",
				GroupID:      "test-group",
				ManifestPath: "/test",
				K8sObject: &unstructured.Unstructured{
					Object: tm{
						"apiVersion": "v1",
						"kind":       "Pod",
						"metadata": tm{
							"name":      "test-name",
							"namespace": "test-ns",
						},
					},
				},
			},
		},

		"A cluster scoped known type should return correctly the resource (ignoring the namespace).": {
			obj: &unstructured.Unstructured{
				Object: tm{
					"apiVersion": "rbac.authorization.k8s.io/v1",
					"kind":       "ClusterRoleBinding",
					"metadata": tm{
						"name":      "test-name",
						"namespace": "test-ns",
					},
				},
			},
			groupID:      "test-group",
			manifestPath: "/test",
			mock: func(m *modelmock.KubernetesDiscoveryClient) {
				m.On("GetServerGroupsAndResources", mock.Anything).Once().Return(nil, testAPIResourceList, nil)
			},
			expResource: model.Resource{
				ID:           "rbac.authorization.k8s.io/v1/ClusterRoleBinding/default/test-name",
				GroupID:      "test-group",
				ManifestPath: "/test",
				K8sObject: &unstructured.Unstructured{
					Object: tm{
						"apiVersion": "rbac.authorization.k8s.io/v1",
						"kind":       "ClusterRoleBinding",
						"metadata": tm{
							"name":      "test-name",
							"namespace": "test-ns",
						},
					},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)

			// Mocks.
			mk := &modelmock.KubernetesDiscoveryClient{}
			test.mock(mk)

			// Prepare and create.
			f, err := model.NewResourceAndGroupFactory(mk, log.Noop)
			require.NoError(err)
			gotResource, err := f.NewResource(test.obj, test.groupID, test.manifestPath)

			// Check.
			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expResource, *gotResource)
			}
		})
	}
}

func TestTestResourceAndGroupFactoryNewGroup(t *testing.T) {
	fourtyTwo := 42

	tests := map[string]struct {
		id       string
		path     string
		config   model.GroupConfig
		expGroup model.Group
	}{
		"A regular group creation should succeed.": {
			id:   "test1",
			path: "tests/test1",
			config: model.GroupConfig{
				Priority: &fourtyTwo,
				HooksConfig: model.GroupHooksConfig{
					Pre:  &model.GroupHookConfigSpec{Cmd: []string{"cmd1"}, Timeout: 555 * time.Millisecond},
					Post: &model.GroupHookConfigSpec{Cmd: []string{"cmd2"}, Timeout: 444 * time.Millisecond},
				},
			},
			expGroup: model.Group{
				ID:       "test1",
				Path:     "tests/test1",
				Priority: 42,
				Hooks: model.GroupHooks{
					Pre:  &model.GroupHookSpec{Cmd: []string{"cmd1"}, Timeout: 555 * time.Millisecond},
					Post: &model.GroupHookSpec{Cmd: []string{"cmd2"}, Timeout: 444 * time.Millisecond},
				},
			},
		},

		"If group doesn't have priority, default priority should be set.": {
			id:   "test1",
			path: "tests/test1",
			config: model.GroupConfig{
				HooksConfig: model.GroupHooksConfig{
					Pre:  &model.GroupHookConfigSpec{Cmd: []string{"cmd1"}, Timeout: 555 * time.Millisecond},
					Post: &model.GroupHookConfigSpec{Cmd: []string{"cmd2"}, Timeout: 444 * time.Millisecond},
				},
			},
			expGroup: model.Group{
				ID:       "test1",
				Path:     "tests/test1",
				Priority: 1000,
				Hooks: model.GroupHooks{
					Pre:  &model.GroupHookSpec{Cmd: []string{"cmd1"}, Timeout: 555 * time.Millisecond},
					Post: &model.GroupHookSpec{Cmd: []string{"cmd2"}, Timeout: 444 * time.Millisecond},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)

			// Prepare and create.
			mk := &modelmock.KubernetesDiscoveryClient{}
			mk.On("GetServerGroupsAndResources", mock.Anything).Once().Return(nil, nil, nil)

			f, err := model.NewResourceAndGroupFactory(mk, log.Noop)
			require.NoError(err)
			gotGroup := f.NewGroup(test.id, test.path, test.config)

			// Check.
			assert.Equal(t, test.expGroup, gotGroup)
		})
	}
}
