// Code generated by mockery v2.3.0. DO NOT EDIT.

package gitmock

import (
	billy "github.com/go-git/go-billy/v5"
	git "github.com/go-git/go-git/v5"

	mock "github.com/stretchr/testify/mock"

	object "github.com/go-git/go-git/v5/plumbing/object"

	plumbing "github.com/go-git/go-git/v5/plumbing"
)

// GoGitRepoClient is an autogenerated mock type for the GoGitRepoClient type
type GoGitRepoClient struct {
	mock.Mock
}

// Checkout provides a mock function with given fields: opts
func (_m *GoGitRepoClient) Checkout(opts *git.CheckoutOptions) error {
	ret := _m.Called(opts)

	var r0 error
	if rf, ok := ret.Get(0).(func(*git.CheckoutOptions) error); ok {
		r0 = rf(opts)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CommitObject provides a mock function with given fields: h
func (_m *GoGitRepoClient) CommitObject(h plumbing.Hash) (*object.Commit, error) {
	ret := _m.Called(h)

	var r0 *object.Commit
	if rf, ok := ret.Get(0).(func(plumbing.Hash) *object.Commit); ok {
		r0 = rf(h)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*object.Commit)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(plumbing.Hash) error); ok {
		r1 = rf(h)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FileSystem provides a mock function with given fields:
func (_m *GoGitRepoClient) FileSystem() (billy.Filesystem, error) {
	ret := _m.Called()

	var r0 billy.Filesystem
	if rf, ok := ret.Get(0).(func() billy.Filesystem); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(billy.Filesystem)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Head provides a mock function with given fields:
func (_m *GoGitRepoClient) Head() (*plumbing.Reference, error) {
	ret := _m.Called()

	var r0 *plumbing.Reference
	if rf, ok := ret.Get(0).(func() *plumbing.Reference); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*plumbing.Reference)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MergeBase provides a mock function with given fields: current, other
func (_m *GoGitRepoClient) MergeBase(current *object.Commit, other *object.Commit) ([]*object.Commit, error) {
	ret := _m.Called(current, other)

	var r0 []*object.Commit
	if rf, ok := ret.Get(0).(func(*object.Commit, *object.Commit) []*object.Commit); ok {
		r0 = rf(current, other)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*object.Commit)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*object.Commit, *object.Commit) error); ok {
		r1 = rf(current, other)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Patch provides a mock function with given fields: current, other
func (_m *GoGitRepoClient) Patch(current *object.Commit, other *object.Commit) (*object.Patch, error) {
	ret := _m.Called(current, other)

	var r0 *object.Patch
	if rf, ok := ret.Get(0).(func(*object.Commit, *object.Commit) *object.Patch); ok {
		r0 = rf(current, other)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*object.Patch)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*object.Commit, *object.Commit) error); ok {
		r1 = rf(current, other)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ResolveRevision provides a mock function with given fields: rev
func (_m *GoGitRepoClient) ResolveRevision(rev plumbing.Revision) (*plumbing.Hash, error) {
	ret := _m.Called(rev)

	var r0 *plumbing.Hash
	if rf, ok := ret.Get(0).(func(plumbing.Revision) *plumbing.Hash); ok {
		r0 = rf(rev)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*plumbing.Hash)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(plumbing.Revision) error); ok {
		r1 = rf(rev)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
