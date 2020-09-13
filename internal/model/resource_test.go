package model_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/slok/kahoy/internal/model"
)

func TestGenResourceID(t *testing.T) {
	// Helper alias for verbosity of unstructured internal maps.
	type tm = map[string]interface{}

	tests := map[string]struct {
		obj   model.K8sObject
		expID string
	}{
		"A regular resource should gen correct ID": {
			expID: "networking.k8s.io/v1beta1/Ingress/test-ns/test-name",
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
		},

		"A resource with default ns should gen correct ID": {
			expID: "networking.k8s.io/v1beta1/Ingress/default/test-name",
			obj: &unstructured.Unstructured{
				Object: tm{
					"apiVersion": "networking.k8s.io/v1beta1",
					"kind":       "Ingress",
					"metadata": tm{
						"name":      "test-name",
						"namespace": "",
					},
				},
			},
		},

		"A resource witho core group should gen correct ID": {
			expID: "core/v1/Pod/test-ns/test-name",
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
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			id := model.GenResourceID(test.obj)
			assert.Equal(t, test.expID, id)
		})
	}
}

func TestNewGroup(t *testing.T) {
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
				WaitConfig: &model.GroupWaitConfig{
					Duration: 555 * time.Millisecond,
				},
			},
			expGroup: model.Group{
				ID:       "test1",
				Path:     "tests/test1",
				Priority: 42,
				Wait: model.GroupWait{
					Duration: 555 * time.Millisecond,
				},
			},
		},

		"If group doesn't have priority, default priority should be set.": {
			id:   "test1",
			path: "tests/test1",
			config: model.GroupConfig{
				WaitConfig: &model.GroupWaitConfig{
					Duration: 555 * time.Millisecond,
				},
			},
			expGroup: model.Group{
				ID:       "test1",
				Path:     "tests/test1",
				Priority: 1000,
				Wait: model.GroupWait{
					Duration: 555 * time.Millisecond,
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			gotGroup := model.NewGroup(test.id, test.path, test.config)
			assert.Equal(t, test.expGroup, gotGroup)
		})
	}
}
