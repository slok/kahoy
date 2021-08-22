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

func newResource(kubeAPIVersion, kubeType, ns, name string) model.Resource {
	return newCustomResource(kubeAPIVersion, kubeType, ns, name, nil, nil)
}

func newLabeledResource(name string, labels map[string]string) model.Resource {
	return newCustomResource("v1", "Pod", "testns", name, labels, nil)
}

func newAnnotatedResource(name string, annotations map[string]string) model.Resource {
	return newCustomResource("v1", "Pod", "testns", name, nil, annotations)
}

func newCustomResource(kubeAPIVersion, kubeType, ns, name string, labels, annotations map[string]string) model.Resource {
	type tm = map[string]interface{}

	objAnnotations := tm{}
	for k, v := range annotations {
		objAnnotations[k] = v
	}
	objLabels := tm{}
	for k, v := range labels {
		objLabels[k] = v
	}
	return model.Resource{
		ID: name,
		K8sObject: &unstructured.Unstructured{
			Object: tm{
				"apiVersion": kubeAPIVersion,
				"kind":       kubeType,
				"metadata": tm{
					"name":        name,
					"namespace":   ns,
					"labels":      objLabels,
					"annotations": objAnnotations,
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

func TestIncludeNamespaceProcessor(t *testing.T) {
	tests := map[string]struct {
		regexes      []string
		resources    []model.Resource
		expResources []model.Resource
		expErr       bool
	}{
		"No regexes should not filter anything.": {
			regexes: []string{},
			resources: []model.Resource{
				newResource("v1", "Pod", "test-ns1", "test-name"),
				newResource("v1", "Pod", "test-ns2", "test-name"),
			},
			expResources: []model.Resource{
				newResource("v1", "Pod", "test-ns1", "test-name"),
				newResource("v1", "Pod", "test-ns2", "test-name"),
			},
		},

		"Matching regexes should keep resources and exclude non matching ones.": {
			regexes: []string{
				"test-ns*",
			},
			resources: []model.Resource{
				newResource("v1", "Pod", "test-ns1", "test-name"),
				newResource("v1", "Pod", "test-ns2", "test-name"),
				newResource("v1", "Pod", "ns3", "test-name"),
			},
			expResources: []model.Resource{
				newResource("v1", "Pod", "test-ns1", "test-name"),
				newResource("v1", "Pod", "test-ns2", "test-name"),
			},
		},

		"If no resources match given regex no resources are returned.": {
			regexes: []string{
				"test-ns*",
			},
			resources: []model.Resource{
				newResource("v1", "Pod", "ns1", "test-name"),
				newResource("v1", "Pod", "ns2", "test-name"),
				newResource("v1", "Pod", "ns3", "test-name"),
			},
			expResources: []model.Resource{},
		},

		"If all resources match regexes, all resources are returned.": {
			regexes: []string{
				"ns1",
				"ns2",
			},
			resources: []model.Resource{
				newResource("v1", "Pod", "ns1", "test-name"),
				newResource("v1", "Pod", "ns2", "test-name"),
			},
			expResources: []model.Resource{
				newResource("v1", "Pod", "ns1", "test-name"),
				newResource("v1", "Pod", "ns2", "test-name"),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			proc, err := process.NewIncludeNamespaceProcessor(test.regexes, log.Noop)
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

func TestKubeLabelSelectorProcessor(t *testing.T) {
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
			selector: "k1=v1,k2 in (v21,v22),k3!=v3",
			resources: []model.Resource{
				newLabeledResource("r0", map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"}),
				newLabeledResource("r1", map[string]string{"k2": "v2", "k3": "v3"}),
				newLabeledResource("r2", map[string]string{"k1": "v1", "k2": "v21", "k4": "v4"}),
				newLabeledResource("r3", map[string]string{"k1": "v1", "k2": "v2", "k3": "v3", "k4": "v4"}),
				newLabeledResource("r4", map[string]string{"k1": "v1", "k4": "v4"}),
				newLabeledResource("r5", map[string]string{"k1": "v1", "k2": "v22", "k5": "v5"}),
				newLabeledResource("r6", map[string]string{"k1": "v1", "k2": "v23", "k5": "v5"}),
			},
			expResources: []model.Resource{
				newLabeledResource("r2", map[string]string{"k1": "v1", "k2": "v21", "k4": "v4"}),
				newLabeledResource("r5", map[string]string{"k1": "v1", "k2": "v22", "k5": "v5"}),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			proc, err := process.NewKubeLabelSelectorProcessor(test.selector, log.Noop)
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

func TestKubeAnnotationSelectorProcessor(t *testing.T) {
	tests := map[string]struct {
		selector     string
		resources    []model.Resource
		expResources []model.Resource
		expErr       bool
	}{
		"No selector should not filter anything.": {
			selector: "",
			resources: []model.Resource{
				newAnnotatedResource("r0", map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"}),
				newAnnotatedResource("r1", map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"}),
			},
			expResources: []model.Resource{
				newAnnotatedResource("r0", map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"}),
				newAnnotatedResource("r1", map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"}),
			},
		},

		"Equality selector should filter the ones that don't have that annotation.": {
			selector: "k1=v1",
			resources: []model.Resource{
				newAnnotatedResource("r0", map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"}),
				newAnnotatedResource("r1", map[string]string{"k2": "v2", "k3": "v3"}),
			},
			expResources: []model.Resource{
				newAnnotatedResource("r0", map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"}),
			},
		},

		"Not equality selector should filter the ones that have that annotation.": {
			selector: "k1!=v1",
			resources: []model.Resource{
				newAnnotatedResource("r0", map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"}),
				newAnnotatedResource("r1", map[string]string{"k2": "v2", "k3": "v3"}),
			},
			expResources: []model.Resource{
				newAnnotatedResource("r1", map[string]string{"k2": "v2", "k3": "v3"}),
			},
		},

		"Multiple selectors should filter the correctly.": {
			selector: "k1=v1,k2 in (v21,v22),k3!=v3",
			resources: []model.Resource{
				newAnnotatedResource("r0", map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"}),
				newAnnotatedResource("r1", map[string]string{"k2": "v2", "k3": "v3"}),
				newAnnotatedResource("r2", map[string]string{"k1": "v1", "k2": "v21", "k4": "v4"}),
				newAnnotatedResource("r3", map[string]string{"k1": "v1", "k2": "v2", "k3": "v3", "k4": "v4"}),
				newAnnotatedResource("r4", map[string]string{"k1": "v1", "k4": "v4"}),
				newAnnotatedResource("r5", map[string]string{"k1": "v1", "k2": "v22", "k5": "v5"}),
				newAnnotatedResource("r6", map[string]string{"k1": "v1", "k2": "v23", "k5": "v5"}),
			},
			expResources: []model.Resource{
				newAnnotatedResource("r2", map[string]string{"k1": "v1", "k2": "v21", "k4": "v4"}),
				newAnnotatedResource("r5", map[string]string{"k1": "v1", "k2": "v22", "k5": "v5"}),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			proc, err := process.NewKubeAnnotationSelectorProcessor(test.selector, log.Noop)
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
