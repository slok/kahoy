// Code generated by mockery v2.3.0. DO NOT EDIT.

package storagemock

import (
	context "context"

	model "github.com/slok/kahoy/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// StateRepository is an autogenerated mock type for the StateRepository type
type StateRepository struct {
	mock.Mock
}

// StoreState provides a mock function with given fields: ctx, state
func (_m *StateRepository) StoreState(ctx context.Context, state model.State) error {
	ret := _m.Called(ctx, state)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, model.State) error); ok {
		r0 = rf(ctx, state)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
