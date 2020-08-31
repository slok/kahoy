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
		oldRes   []model.Resource
		newRes   []model.Resource
		expState []plan.State
		expErr   error
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
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			p := plan.NewPlanner(false, log.Noop)
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
