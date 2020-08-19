package process

import (
	"context"

	"github.com/slok/kahoy/internal/model"
)

// ResourceProcessor knows how to process a group of resources.
type ResourceProcessor interface {
	Process(ctx context.Context, resources []model.Resource) ([]model.Resource, error)
}

//go:generate mockery --case underscore --output processmock --outpkg processmock --name ResourceProcessor

// ResourceProcessorFunc is a helper type to create processors without the need to create new types.
type ResourceProcessorFunc func(ctx context.Context, resources []model.Resource) ([]model.Resource, error)

// Interface assertion.
var _ ResourceProcessor = ResourceProcessorFunc(nil)

// Process satisfies ResourceProcessor interface.
func (f ResourceProcessorFunc) Process(ctx context.Context, resources []model.Resource) ([]model.Resource, error) {
	return f(ctx, resources)
}

// NewResourceProcessorChain returns a chain of processors that will execute all the processors available
// in the chain passing the result of one processor as the parameter to the next until the chain ends or
// one of the processors fail.
func NewResourceProcessorChain(processors ...ResourceProcessor) ResourceProcessor {
	return ResourceProcessorFunc(func(ctx context.Context, resources []model.Resource) ([]model.Resource, error) {
		var err error
		for _, proc := range processors {
			resources, err = proc.Process(ctx, resources)
			if err != nil {
				return resources, err
			}
		}

		return resources, nil
	})
}
