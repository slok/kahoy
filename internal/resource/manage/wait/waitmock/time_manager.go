// Code generated by mockery v2.1.0. DO NOT EDIT.

package waitmock

import (
	context "context"
	time "time"

	mock "github.com/stretchr/testify/mock"
)

// TimeManager is an autogenerated mock type for the TimeManager type
type TimeManager struct {
	mock.Mock
}

// Sleep provides a mock function with given fields: ctx, d
func (_m *TimeManager) Sleep(ctx context.Context, d time.Duration) {
	_m.Called(ctx, d)
}