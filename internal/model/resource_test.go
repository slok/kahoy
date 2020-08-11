package model_test

import (
	"testing"

	"github.com/slok/kahoy/internal/model"
	"gotest.tools/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
