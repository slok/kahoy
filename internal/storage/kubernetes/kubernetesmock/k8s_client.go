// Code generated by mockery (devel). DO NOT EDIT.

package kubernetesmock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	v1 "k8s.io/api/core/v1"
)

// K8sClient is an autogenerated mock type for the K8sClient type
type K8sClient struct {
	mock.Mock
}

// EnsureMissingSecret provides a mock function with given fields: ctx, ns, name
func (_m *K8sClient) EnsureMissingSecret(ctx context.Context, ns string, name string) error {
	ret := _m.Called(ctx, ns, name)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, ns, name)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// EnsureSecret provides a mock function with given fields: ctx, sec
func (_m *K8sClient) EnsureSecret(ctx context.Context, sec *v1.Secret) error {
	ret := _m.Called(ctx, sec)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1.Secret) error); ok {
		r0 = rf(ctx, sec)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetSecret provides a mock function with given fields: ctx, ns, name
func (_m *K8sClient) GetSecret(ctx context.Context, ns string, name string) (*v1.Secret, error) {
	ret := _m.Called(ctx, ns, name)

	var r0 *v1.Secret
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *v1.Secret); ok {
		r0 = rf(ctx, ns, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.Secret)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, ns, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListSecrets provides a mock function with given fields: ctx, ns, labelFilter
func (_m *K8sClient) ListSecrets(ctx context.Context, ns string, labelFilter map[string]string) ([]v1.Secret, error) {
	ret := _m.Called(ctx, ns, labelFilter)

	var r0 []v1.Secret
	if rf, ok := ret.Get(0).(func(context.Context, string, map[string]string) []v1.Secret); ok {
		r0 = rf(ctx, ns, labelFilter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]v1.Secret)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, map[string]string) error); ok {
		r1 = rf(ctx, ns, labelFilter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
