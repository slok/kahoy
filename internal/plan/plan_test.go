package plan_test

import (
	"context"
	"errors"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/plan"
)

func TestPlannerPlan(t *testing.T) {
	tests := map[string]struct {
		currentRes  []model.Resource
		expectedRes []model.Resource
		expState    []plan.State
		expErr      error
	}{
		"Without current and expected resources, should plan empty list of states.": {
			expState: []plan.State{},
		},

		"Without current and with expected resources, should plan list withouth missing states.": {
			currentRes: []model.Resource{},
			expectedRes: []model.Resource{
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

		"With same current and expected resources, should plan list withouth missing states.": {
			currentRes: []model.Resource{
				{ID: "test0"},
				{ID: "test1"},
				{ID: "test2"},
				{ID: "test3"},
			},
			expectedRes: []model.Resource{
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

		"With deleted in the expected resources, should plan list with missing states.": {
			currentRes: []model.Resource{
				{ID: "test0"},
				{ID: "test1"},
				{ID: "test2"},
				{ID: "test3"},
			},
			expectedRes: []model.Resource{},
			expState: []plan.State{
				{Resource: model.Resource{ID: "test0"}, State: plan.ResourceStateMissing},
				{Resource: model.Resource{ID: "test1"}, State: plan.ResourceStateMissing},
				{Resource: model.Resource{ID: "test2"}, State: plan.ResourceStateMissing},
				{Resource: model.Resource{ID: "test3"}, State: plan.ResourceStateMissing},
			},
		},

		"With some deleted and some new in the expected resources, should plan list with missing states.": {
			currentRes: []model.Resource{
				{ID: "test0"},
				{ID: "test1"},
				{ID: "test2"},
				{ID: "test3"},
			},
			expectedRes: []model.Resource{
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
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			p := plan.NewPlanner(log.Noop)
			gotState, err := p.Plan(context.TODO(), test.expectedRes, test.currentRes)

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
