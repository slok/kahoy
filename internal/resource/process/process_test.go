package process_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/resource/process"
	"github.com/slok/kahoy/internal/resource/process/processmock"
)

func TestResourceProcessorChain(t *testing.T) {
	tests := map[string]struct {
		processorMocks func() []*processmock.ResourceProcessor
		res            []model.Resource
		expResources   []model.Resource
		expErr         bool
	}{
		"No processors, should not fail.": {
			processorMocks: func() []*processmock.ResourceProcessor { return nil },
			res: []model.Resource{
				{ID: "test1"},
			},
			expResources: []model.Resource{
				{ID: "test1"},
			},
		},

		"Having multiple processors should call them and pass the results on one as the argument of the next.": {
			processorMocks: func() []*processmock.ResourceProcessor {
				m1 := &processmock.ResourceProcessor{}
				m2 := &processmock.ResourceProcessor{}
				m3 := &processmock.ResourceProcessor{}

				m1.On("Process", mock.Anything, []model.Resource{{ID: "test1"}}).Once().Return([]model.Resource{{ID: "test2"}}, nil)
				m2.On("Process", mock.Anything, []model.Resource{{ID: "test2"}}).Once().Return([]model.Resource{{ID: "test3"}}, nil)
				m3.On("Process", mock.Anything, []model.Resource{{ID: "test3"}}).Once().Return([]model.Resource{{ID: "test4"}}, nil)

				return []*processmock.ResourceProcessor{m1, m2, m3}
			},
			res: []model.Resource{
				{ID: "test1"},
			},
			expResources: []model.Resource{
				{ID: "test4"},
			},
		},

		"Having multiple processors, if one fails it should stop the chain and return the latest correct resources.": {
			processorMocks: func() []*processmock.ResourceProcessor {
				m1 := &processmock.ResourceProcessor{}
				m2 := &processmock.ResourceProcessor{}
				m3 := &processmock.ResourceProcessor{}

				m1.On("Process", mock.Anything, mock.Anything).Once().Return(nil, nil)
				m2.On("Process", mock.Anything, mock.Anything).Once().Return([]model.Resource{{ID: "test2"}}, errors.New("whatever"))

				return []*processmock.ResourceProcessor{m1, m2, m3}
			},
			res: []model.Resource{
				{ID: "test1"},
			},
			expResources: []model.Resource{
				{ID: "test2"},
			},
			expErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			// Mocks.
			mocks := test.processorMocks()
			procs := make([]process.ResourceProcessor, 0, len(mocks))
			for _, m := range mocks {
				procs = append(procs, m)
			}

			// Prepare and execute.
			proc := process.NewResourceProcessorChain(procs...)
			gotResources, err := proc.Process(context.TODO(), test.res)

			// Check.
			assert.Equal(test.expErr, err != nil)
			assert.Equal(test.expResources, gotResources)
			for _, m := range mocks {
				m.AssertExpectations(t)
			}
		})
	}
}
