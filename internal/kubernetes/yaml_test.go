package kubernetes_test

import (
	"context"
	"sort"
	"testing"

	"github.com/slok/kahoy/internal/kubernetes"
	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Helper alias for verbosity of unstructured internal maps.
type (
	tm = map[string]interface{}
	ts = []interface{}
)

func TestYAMLObjectSerializerDecodeObjects(t *testing.T) {

	tests := map[string]struct {
		rawObjects string
		expObjs    []model.K8sObject
		expErr     bool
	}{
		"Invalid YAML should error.": {
			rawObjects: "()",
			expErr:     true,
		},

		"No objects on file should return no objects.": {
			rawObjects: "",
			expObjs:    []model.K8sObject{},
		},

		"Empty objects on file should return no objects.": {
			rawObjects: `---

---
---

---


---`,
			expObjs: []model.K8sObject{},
		},
		"YAML comments should be treated as  empty yaml files.": {
			rawObjects: `# some comments
---
# other comment
---
# multiline
# comment
---

---`,
			expObjs: []model.K8sObject{},
		},

		"Deconding single object should return the decoded object.": {
			rawObjects: `---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-name
  namespace: test-ns
  labels:
    l1: v1
    l2: v2
data:
  k1: v1
  k2: v2
`,
			expObjs: []model.K8sObject{
				&unstructured.Unstructured{
					Object: tm{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": tm{
							"name":      "test-name",
							"namespace": "test-ns",
							"labels": tm{
								"l1": "v1",
								"l2": "v2",
							},
						},
						"data": tm{
							"k1": "v1",
							"k2": "v2",
						},
					},
				},
			},
		},

		"Deconding multiple object should return the decoded objects.": {
			rawObjects: `---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-name
  namespace: test-ns
  labels:
    l1: v1
    l2: v2
data:
  k1: v1
  k2: v2

---
kind: Service
apiVersion: v1
metadata:
  name: test2-name
  namespace: test2-ns
  labels:
    l21: v21
    l22: v22
spec:
  selector:
    l21: v21
  type: ClusterIP
  ports:
    - name: http
      port: 8080
`,
			expObjs: []model.K8sObject{
				&unstructured.Unstructured{
					Object: tm{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": tm{
							"name":      "test-name",
							"namespace": "test-ns",
							"labels": tm{
								"l1": "v1",
								"l2": "v2",
							},
						},
						"data": tm{
							"k1": "v1",
							"k2": "v2",
						},
					},
				},
				&unstructured.Unstructured{
					Object: tm{
						"apiVersion": "v1",
						"kind":       "Service",
						"metadata": tm{
							"name":      "test2-name",
							"namespace": "test2-ns",
							"labels": tm{
								"l21": "v21",
								"l22": "v22",
							},
						},
						"spec": tm{
							"selector": tm{
								"l21": "v21",
							},
							"type": "ClusterIP",
							"ports": ts{
								tm{
									"name": "http",
									"port": int64(8080),
								},
							},
						},
					},
				},
			},
		},

		"Deconding multiple object in a list object should return the decoded objects.": {
			rawObjects: `
apiVersion: v1
kind: List
items:
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: test-name
    namespace: test-ns
    labels:
      l1: v1
      l2: v2
  data:
    k1: v1
    k2: v2
- apiVersion: v1
  kind: Service
  metadata:
    name: test2-name
    namespace: test2-ns
    labels:
      l21: v21
      l22: v22
  spec:
    selector:
      l21: v21
    type: ClusterIP
    ports:
      - name: http
        port: 8080
metadata:
  resourceVersion: ""
  selfLink: ""
`,
			expObjs: []model.K8sObject{
				&unstructured.Unstructured{
					Object: tm{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": tm{
							"name":      "test-name",
							"namespace": "test-ns",
							"labels": tm{
								"l1": "v1",
								"l2": "v2",
							},
						},
						"data": tm{
							"k1": "v1",
							"k2": "v2",
						},
					},
				},
				&unstructured.Unstructured{
					Object: tm{
						"apiVersion": "v1",
						"kind":       "Service",
						"metadata": tm{
							"name":      "test2-name",
							"namespace": "test2-ns",
							"labels": tm{
								"l21": "v21",
								"l22": "v22",
							},
						},
						"spec": tm{
							"selector": tm{
								"l21": "v21",
							},
							"type": "ClusterIP",
							"ports": ts{
								tm{
									"name": "http",
									"port": int64(8080),
								},
							},
						},
					},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			loader := kubernetes.NewYAMLObjectSerializer(log.Noop)
			gotObjs, err := loader.DecodeObjects(context.TODO(), []byte(test.rawObjects))

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				// Sort for reliable comparison.
				sort.SliceStable(test.expObjs, func(i, j int) bool { return test.expObjs[i].GetName() < test.expObjs[j].GetName() })
				sort.SliceStable(gotObjs, func(i, j int) bool { return gotObjs[i].GetName() < gotObjs[j].GetName() })

				assert.Equal(test.expObjs, gotObjs)
			}
		})
	}
}

func TestYAMLObjectSerializerEncodeObjects(t *testing.T) {
	tests := map[string]struct {
		objs   []model.K8sObject
		expRaw string
		expErr bool
	}{
		"No objects should return empty raw.": {
			objs:   []model.K8sObject{},
			expRaw: "",
		},

		"Nil objects should be ignored.": {
			objs:   []model.K8sObject{nil},
			expRaw: "",
		},

		"Single object should be encoded as a raw.": {
			objs: []model.K8sObject{
				&unstructured.Unstructured{
					Object: tm{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": tm{
							"name":      "test-name",
							"namespace": "test-ns",
							"labels": tm{
								"l1": "v1",
								"l2": "v2",
							},
						},
						"data": tm{
							"k1": "v1",
							"k2": "v2",
						},
					},
				},
			},
			expRaw: `---
apiVersion: v1
data:
  k1: v1
  k2: v2
kind: ConfigMap
metadata:
  labels:
    l1: v1
    l2: v2
  name: test-name
  namespace: test-ns
`,
		},

		"Multiple objects should be encoded as raw.": {
			objs: []model.K8sObject{
				&unstructured.Unstructured{
					Object: tm{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": tm{
							"name":      "test-name",
							"namespace": "test-ns",
							"labels": tm{
								"l1": "v1",
								"l2": "v2",
							},
						},
						"data": tm{
							"k1": "v1",
							"k2": "v2",
						},
					},
				},
				&unstructured.Unstructured{
					Object: tm{
						"apiVersion": "v1",
						"kind":       "Service",
						"metadata": tm{
							"name":      "test2-name",
							"namespace": "test2-ns",
							"labels": tm{
								"l21": "v21",
								"l22": "v22",
							},
						},
						"spec": tm{
							"selector": tm{
								"l21": "v21",
							},
							"type": "ClusterIP",
							"ports": ts{
								tm{
									"name": "http",
									"port": int64(8080),
								},
							},
						},
					},
				},
			},
			expRaw: `---
apiVersion: v1
data:
  k1: v1
  k2: v2
kind: ConfigMap
metadata:
  labels:
    l1: v1
    l2: v2
  name: test-name
  namespace: test-ns
---
apiVersion: v1
kind: Service
metadata:
  labels:
    l21: v21
    l22: v22
  name: test2-name
  namespace: test2-ns
spec:
  ports:
  - name: http
    port: 8080
  selector:
    l21: v21
  type: ClusterIP
`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			loader := kubernetes.NewYAMLObjectSerializer(log.Noop)
			gotRaw, err := loader.EncodeObjects(context.TODO(), test.objs)

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expRaw, string(gotRaw))
			}
		})
	}
}
