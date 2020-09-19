package json_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/storage/json"
)

func newCustomResource(kAPIVersion, kType, ns, name, group string) model.Resource {
	type tm = map[string]interface{}

	return model.Resource{
		ID:      name,
		GroupID: group,
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

func TestStateRepository(t *testing.T) {
	t0, _ := time.Parse(time.RFC3339, "1912-06-23T01:02:03Z")
	t1, _ := time.Parse(time.RFC3339, "1912-06-23T01:02:42Z")

	tests := map[string]struct {
		state  model.State
		expOut string
		expErr bool
	}{
		"Having no resources should give the correct state without resorces": {
			state: model.State{
				ID:        "id1",
				StartedAt: t0,
				EndedAt:   t1,
			},
			expOut: `{"version":"v1","id":"id1","started_at":"1912-06-23T01:02:03Z","ended_at":"1912-06-23T01:02:42Z","applied_resources":[],"deleted_resources":[]}`,
		},

		"Having resources should give the correct state without resorces": {
			state: model.State{
				ID:        "id1",
				StartedAt: t0,
				EndedAt:   t1,
				AppliedResources: []model.Resource{
					newCustomResource("v1", "Pod", "ns1", "applied1", "group1"),
					newCustomResource("networking.k8s.io/v1beta1", "Ingress", "ns2", "applied2", "group2"),
				},
				DeletedResources: []model.Resource{
					newCustomResource("apps/v1", "Deployment", "ns3", "applied3", "group3"),
					newCustomResource("rbac.authorization.k8s.io/v1", "Role", "ns4", "applied4", "group4"),
				},
			},
			expOut: `{"version":"v1","id":"id1","started_at":"1912-06-23T01:02:03Z","ended_at":"1912-06-23T01:02:42Z","applied_resources":[{"id":"applied1","group":"group1","gvk":"/v1/Pod","api_version":"v1","kind":"Pod","namespace":"ns1","name":"applied1"},{"id":"applied2","group":"group2","gvk":"networking.k8s.io/v1beta1/Ingress","api_version":"networking.k8s.io/v1beta1","kind":"Ingress","namespace":"ns2","name":"applied2"}],"deleted_resources":[{"id":"applied3","group":"group3","gvk":"apps/v1/Deployment","api_version":"apps/v1","kind":"Deployment","namespace":"ns3","name":"applied3"},{"id":"applied4","group":"group4","gvk":"rbac.authorization.k8s.io/v1/Role","api_version":"rbac.authorization.k8s.io/v1","kind":"Role","namespace":"ns4","name":"applied4"}]}`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			var out bytes.Buffer
			r := json.NewStateRepository(&out)
			err := r.StoreState(context.TODO(), test.state)

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				gotOut := out.String()
				assert.Equal(test.expOut, gotOut)
			}
		})
	}
}
