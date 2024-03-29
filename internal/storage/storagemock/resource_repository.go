// Code generated by mockery (devel). DO NOT EDIT.

package storagemock

import (
	context "context"

	model "github.com/slok/kahoy/internal/model"
	mock "github.com/stretchr/testify/mock"

	storage "github.com/slok/kahoy/internal/storage"
)

// ResourceRepository is an autogenerated mock type for the ResourceRepository type
type ResourceRepository struct {
	mock.Mock
}

// GetResource provides a mock function with given fields: ctx, id
func (_m *ResourceRepository) GetResource(ctx context.Context, id string) (*model.Resource, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.Resource
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.Resource); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Resource)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListResources provides a mock function with given fields: ctx, opts
func (_m *ResourceRepository) ListResources(ctx context.Context, opts storage.ResourceListOpts) (*storage.ResourceList, error) {
	ret := _m.Called(ctx, opts)

	var r0 *storage.ResourceList
	if rf, ok := ret.Get(0).(func(context.Context, storage.ResourceListOpts) *storage.ResourceList); ok {
		r0 = rf(ctx, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*storage.ResourceList)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, storage.ResourceListOpts) error); ok {
		r1 = rf(ctx, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
