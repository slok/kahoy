package process_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/resource/process"
)

func newResource(kAPIVersion, kType, ns, name string) model.Resource {
	type tm = map[string]interface{}

	return model.Resource{
		K8sObject: &unstructured.Unstructured{
			Object: tm{
				"apiVersion": kAPIVersion,
				"kind":       kType,
				"metadata": tm{
					"name":      name,
					"namespace": ns,
				},
			},
		},
	}
}

func TestExcludeKubeType(t *testing.T) {
	tests := map[string]struct {
		regexes      []string
		resources    []model.Resource
		expResources []model.Resource
		expErr       bool
	}{
		"No regexes should not filter anything.": {
			regexes: []string{},
			resources: []model.Resource{
				newResource("v1", "Pod", "test-ns", "test-name"),
				newResource("networking.k8s.io/v1beta1", "Ingress", "test-ns", "test-name2"),
			},
			expResources: []model.Resource{
				newResource("v1", "Pod", "test-ns", "test-name"),
				newResource("networking.k8s.io/v1beta1", "Ingress", "test-ns", "test-name2"),
			},
		},

		"Not mathing regexes should not filter resources.": {
			regexes: []string{
				"Deployment",
			},
			resources: []model.Resource{
				newResource("v1", "Pod", "test-ns", "test-name"),
				newResource("networking.k8s.io/v1beta1", "Ingress", "test-ns", "test-name2"),
			},
			expResources: []model.Resource{
				newResource("v1", "Pod", "test-ns", "test-name"),
				newResource("networking.k8s.io/v1beta1", "Ingress", "test-ns", "test-name2"),
			},
		},

		"Mathing regexes should filter resources (specific type).": {
			regexes: []string{
				".*/Pod",
			},
			resources: []model.Resource{
				newResource("v1", "Pod", "test-ns", "test-name"),
				newResource("networking.k8s.io/v1beta1", "Ingress", "test-ns", "test-name2"),
				newResource("v1", "Pod", "test-ns", "test-name3"),
				newResource("v1", "Service", "test-ns", "test-name4"),
			},
			expResources: []model.Resource{
				newResource("networking.k8s.io/v1beta1", "Ingress", "test-ns", "test-name2"),
				newResource("v1", "Service", "test-ns", "test-name4"),
			},
		},

		"Mathing regexes should filter resources (specific group).": {
			regexes: []string{
				"^v1/.*",
			},
			resources: []model.Resource{
				newResource("v1", "Pod", "test-ns", "test-name"),
				newResource("networking.k8s.io/v1beta1", "Ingress", "test-ns", "test-name2"),
				newResource("v1", "Pod", "test-ns", "test-name3"),
				newResource("v1", "Service", "test-ns", "test-name4"),
			},
			expResources: []model.Resource{
				newResource("networking.k8s.io/v1beta1", "Ingress", "test-ns", "test-name2"),
			},
		},

		"Mathing regexes should filter resources (specific type and group).": {
			regexes: []string{
				"^apps/v1/StatefulSet$",
			},
			resources: []model.Resource{
				newResource("apps/v1", "StatefulSet", "test-ns", "test-name"),
				newResource("apps/v1", "StatefulSet", "test-ns", "test-name2"),
				newResource("apps/v1", "Deployment", "test-ns", "test-name3"),
				newResource("apps/v1", "Deployment", "test-ns", "test-name4"),
			},
			expResources: []model.Resource{
				newResource("apps/v1", "Deployment", "test-ns", "test-name3"),
				newResource("apps/v1", "Deployment", "test-ns", "test-name4"),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			proc, err := process.NewExcludeKubeTypeProcessor(test.regexes, log.Noop)
			require.NoError(err)
			gotResources, err := proc.Process(context.TODO(), test.resources)

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expResources, gotResources)
			}
		})
	}
}
