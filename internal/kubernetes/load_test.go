package kubernetes_test

import (
	"context"
	"testing"

	"github.com/slok/kahoy/internal/kubernetes"
	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestYAMLObjectLoaderLoadObjects(t *testing.T) {
	// Helper alias for verbosity of unstructured internal maps.
	type (
		tm = map[string]interface{}
		ts = []interface{}
	)

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

		"Deconding multiple object should return the decoded object.": {
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
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			loader := kubernetes.NewYAMLObjectLoader(log.Noop)
			gotObjs, err := loader.LoadObjects(context.TODO(), []byte(test.rawObjects))

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expObjs, gotObjs)
			}
		})
	}
}
