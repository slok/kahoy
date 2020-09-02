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

type tm = map[string]interface{}

func newResource(kAPIVersion, kType, ns, name string) model.Resource {
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

func newLabeledResource(name string, labels map[string]string) model.Resource {
	objLabels := tm{}
	for k, v := range labels {
		objLabels[k] = v
	}
	return model.Resource{
		ID: name,
		K8sObject: &unstructured.Unstructured{
			Object: tm{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": tm{
					"name":   name,
					"labels": objLabels,
				},
			},
		},
	}
}

func TestExcludeKubeTypeProcessor(t *testing.T) {
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

func TestKubeSelectorProcessor(t *testing.T) {
	tests := map[string]struct {
		selector     string
		resources    []model.Resource
		expResources []model.Resource
		expErr       bool
	}{
		"No selector should not filter anything.": {
			selector: "",
			resources: []model.Resource{
				newLabeledResource("r0", map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"}),
				newLabeledResource("r1", map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"}),
			},
			expResources: []model.Resource{
				newLabeledResource("r0", map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"}),
				newLabeledResource("r1", map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"}),
			},
		},

		"Equality selector should filter the ones that don't have that label.": {
			selector: "k1=v1",
			resources: []model.Resource{
				newLabeledResource("r0", map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"}),
				newLabeledResource("r1", map[string]string{"k2": "v2", "k3": "v3"}),
			},
			expResources: []model.Resource{
				newLabeledResource("r0", map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"}),
			},
		},

		"Not equality selector should filter the ones that have that label.": {
			selector: "k1!=v1",
			resources: []model.Resource{
				newLabeledResource("r0", map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"}),
				newLabeledResource("r1", map[string]string{"k2": "v2", "k3": "v3"}),
			},
			expResources: []model.Resource{
				newLabeledResource("r1", map[string]string{"k2": "v2", "k3": "v3"}),
			},
		},

		"Multiple selectors should filter the correctly.": {
			selector: "k1=v1,k2=v2,k3!=v3",
			resources: []model.Resource{
				newLabeledResource("r0", map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"}),
				newLabeledResource("r1", map[string]string{"k2": "v2", "k3": "v3"}),
				newLabeledResource("r2", map[string]string{"k1": "v1", "k2": "v2", "k4": "v4"}),
				newLabeledResource("r3", map[string]string{"k1": "v1", "k2": "v2", "k3": "v3", "k4": "v4"}),
				newLabeledResource("r4", map[string]string{"k1": "v1", "k4": "v4"}),
				newLabeledResource("r5", map[string]string{"k1": "v1", "k2": "v2", "k5": "v5"}),
			},
			expResources: []model.Resource{
				newLabeledResource("r2", map[string]string{"k1": "v1", "k2": "v2", "k4": "v4"}),
				newLabeledResource("r5", map[string]string{"k1": "v1", "k2": "v2", "k5": "v5"}),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			proc, err := process.NewKubeSelectorProcessor(test.selector, log.Noop)
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
