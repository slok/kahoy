package plan_test

import (
	"context"
	"errors"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/plan"
)

func newPod(name string, containerNames []string) model.K8sObject {
	type tm = map[string]interface{}

	containers := []tm{}
	for _, cName := range containerNames {
		containers = append(containers, tm{
			"name": cName,
		})
	}

	return &unstructured.Unstructured{
		Object: tm{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": tm{
				"name":      name,
				"namespace": "test",
			},
			"spec": tm{
				"containers": containers,
			},
		},
	}
}

func TestPlannerPlan(t *testing.T) {
	tests := map[string]struct {
		onlyOnDiff bool
		oldRes     []model.Resource
		newRes     []model.Resource
		expState   []plan.State
		expErr     error
	}{
		"Without old and new resources, should plan empty list of states.": {
			expState: []plan.State{},
		},

		"Without old and with new resources, should plan list withouth missing states.": {
			oldRes: []model.Resource{},
			newRes: []model.Resource{
				{ID: "test0"},
				{ID: "test1"},
				{ID: "test2"},
				{ID: "test3"},
			},
			expState: []plan.State{
				{Resource: model.Resource{ID: "test0"}, State: plan.ResourceStateExists},
				{Resource: model.Resource{ID: "test1"}, State: plan.ResourceStateExists},
				{Resource: model.Resource{ID: "test2"}, State: plan.ResourceStateExists},
				{Resource: model.Resource{ID: "test3"}, State: plan.ResourceStateExists},
			},
		},

		"Without old and with new resources, using only diff changes flag, should plan list withouth missing states.": {
			onlyOnDiff: true,
			oldRes:     []model.Resource{},
			newRes: []model.Resource{
				{ID: "test0"},
				{ID: "test1"},
				{ID: "test2"},
				{ID: "test3"},
			},
			expState: []plan.State{
				{Resource: model.Resource{ID: "test0"}, State: plan.ResourceStateExists},
				{Resource: model.Resource{ID: "test1"}, State: plan.ResourceStateExists},
				{Resource: model.Resource{ID: "test2"}, State: plan.ResourceStateExists},
				{Resource: model.Resource{ID: "test3"}, State: plan.ResourceStateExists},
			},
		},

		"With same old and new resources, should plan list withouth missing states.": {
			oldRes: []model.Resource{
				{ID: "test0"},
				{ID: "test1"},
				{ID: "test2"},
				{ID: "test3"},
			},
			newRes: []model.Resource{
				{ID: "test0"},
				{ID: "test1"},
				{ID: "test2"},
				{ID: "test3"},
			},
			expState: []plan.State{
				{Resource: model.Resource{ID: "test0"}, State: plan.ResourceStateExists},
				{Resource: model.Resource{ID: "test1"}, State: plan.ResourceStateExists},
				{Resource: model.Resource{ID: "test2"}, State: plan.ResourceStateExists},
				{Resource: model.Resource{ID: "test3"}, State: plan.ResourceStateExists},
			},
		},

		"With same old and new resources, using only diff changes flag, should plan list with onthe the resources that changed.": {
			onlyOnDiff: true,
			oldRes: []model.Resource{
				{ID: "test0", K8sObject: newPod("test0", []string{"c1", "c2", "c3"})},
				{ID: "test1", K8sObject: newPod("test1", []string{"c1", "c2", "c3"})},
				{ID: "test2", K8sObject: newPod("test2", []string{"c1", "c2", "c3"})},
				{ID: "test3", K8sObject: newPod("test3", []string{"c1", "c2", "c3"})},
				{ID: "test4", K8sObject: newPod("test4", []string{"c1", "c2", "c3"})},
				{ID: "test5", K8sObject: newPod("test5", []string{"c1", "c2", "c3"})},
			},
			newRes: []model.Resource{
				{ID: "test0", K8sObject: newPod("test0", nil)},
				{ID: "test1", K8sObject: newPod("test1", []string{"c2", "c3"})},
				{ID: "test2", K8sObject: newPod("test2", []string{"c1", "c2", "c3"})},
				{ID: "test3", K8sObject: newPod("test3", []string{"c1", "c2", "c3", "c4"})},
				{ID: "test4", K8sObject: newPod("test4", []string{"c3", "c2", "c1"})},
				{ID: "test5", K8sObject: newPod("test5", []string{"c1", "c2", "c3"})},
			},
			expState: []plan.State{
				{Resource: model.Resource{ID: "test0", K8sObject: newPod("test0", nil)}, State: plan.ResourceStateExists},
				{Resource: model.Resource{ID: "test1", K8sObject: newPod("test1", []string{"c2", "c3"})}, State: plan.ResourceStateExists},
				{Resource: model.Resource{ID: "test3", K8sObject: newPod("test3", []string{"c1", "c2", "c3", "c4"})}, State: plan.ResourceStateExists},
				{Resource: model.Resource{ID: "test4", K8sObject: newPod("test4", []string{"c3", "c2", "c1"})}, State: plan.ResourceStateExists},
			},
		},

		"With deleted in the new resources, should plan list with missing states.": {
			oldRes: []model.Resource{
				{ID: "test0"},
				{ID: "test1"},
				{ID: "test2"},
				{ID: "test3"},
			},
			newRes: []model.Resource{},
			expState: []plan.State{
				{Resource: model.Resource{ID: "test0"}, State: plan.ResourceStateMissing},
				{Resource: model.Resource{ID: "test1"}, State: plan.ResourceStateMissing},
				{Resource: model.Resource{ID: "test2"}, State: plan.ResourceStateMissing},
				{Resource: model.Resource{ID: "test3"}, State: plan.ResourceStateMissing},
			},
		},

		"With deleted in the new resources, using only diff changes flag, should plan list with missing states.": {
			onlyOnDiff: true,
			oldRes: []model.Resource{
				{ID: "test0"},
				{ID: "test1"},
				{ID: "test2"},
				{ID: "test3"},
			},
			newRes: []model.Resource{},
			expState: []plan.State{
				{Resource: model.Resource{ID: "test0"}, State: plan.ResourceStateMissing},
				{Resource: model.Resource{ID: "test1"}, State: plan.ResourceStateMissing},
				{Resource: model.Resource{ID: "test2"}, State: plan.ResourceStateMissing},
				{Resource: model.Resource{ID: "test3"}, State: plan.ResourceStateMissing},
			},
		},

		"With some deleted and some new in the new resources, should plan list with missing states.": {
			oldRes: []model.Resource{
				{ID: "test0"},
				{ID: "test1"},
				{ID: "test2"},
				{ID: "test3"},
			},
			newRes: []model.Resource{
				{ID: "test0"},
				{ID: "test2"},
				{ID: "test4"},
			},
			expState: []plan.State{
				{Resource: model.Resource{ID: "test0"}, State: plan.ResourceStateExists},
				{Resource: model.Resource{ID: "test1"}, State: plan.ResourceStateMissing},
				{Resource: model.Resource{ID: "test2"}, State: plan.ResourceStateExists},
				{Resource: model.Resource{ID: "test3"}, State: plan.ResourceStateMissing},
				{Resource: model.Resource{ID: "test4"}, State: plan.ResourceStateExists},
			},
		},

		"With some deleted and some new in the new resources, using only diff changes flag, should plan list with missing states and changes.": {
			onlyOnDiff: true,
			oldRes: []model.Resource{
				{ID: "test0", K8sObject: newPod("test0", []string{"c1", "c2", "c3"})},
				{ID: "test1", K8sObject: newPod("test1", []string{"c1", "c2", "c3"})},
				{ID: "test2", K8sObject: newPod("test2", []string{"c1", "c2", "c3"})},
				{ID: "test3", K8sObject: newPod("test3", []string{"c1", "c2", "c3"})},
				{ID: "test4", K8sObject: newPod("test4", []string{"c1", "c2", "c3"})},
				{ID: "test5", K8sObject: newPod("test5", []string{"c1", "c2", "c3"})},
			},
			newRes: []model.Resource{
				{ID: "test0", K8sObject: newPod("test0", []string{"c1", "c2", "c3", "c4"})},
				{ID: "test2", K8sObject: newPod("test2", []string{"c1", "c2", "c3"})},
				{ID: "test4", K8sObject: newPod("test4", []string{"c1", "c2", "c3"})},
				{ID: "test5", K8sObject: newPod("test5", []string{"c3", "c1", "c2"})},
			},
			expState: []plan.State{
				{Resource: model.Resource{ID: "test0", K8sObject: newPod("test0", []string{"c1", "c2", "c3", "c4"})}, State: plan.ResourceStateExists},
				{Resource: model.Resource{ID: "test1", K8sObject: newPod("test1", []string{"c1", "c2", "c3"})}, State: plan.ResourceStateMissing},
				{Resource: model.Resource{ID: "test3", K8sObject: newPod("test3", []string{"c1", "c2", "c3"})}, State: plan.ResourceStateMissing},
				{Resource: model.Resource{ID: "test5", K8sObject: newPod("test5", []string{"c3", "c1", "c2"})}, State: plan.ResourceStateExists},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			p := plan.NewPlanner(test.onlyOnDiff, log.Noop)
			gotState, err := p.Plan(context.TODO(), test.oldRes, test.newRes)

			if test.expErr != nil {
				assert.True(errors.Is(err, test.expErr))
			} else if assert.NoError(err) {
				sortStateList(test.expState)
				sortStateList(gotState)
				assert.Equal(test.expState, gotState)
			}
		})
	}
}

func sortStateList(l []plan.State) {
	sort.SliceStable(l, func(i, j int) bool {
		return l[i].Resource.ID < l[j].Resource.ID
	})
}
