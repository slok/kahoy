// Code generated by mockery (devel). DO NOT EDIT.

package fsmock

import (
	filepath "path/filepath"

	mock "github.com/stretchr/testify/mock"
)

// FileSystemManager is an autogenerated mock type for the FileSystemManager type
type FileSystemManager struct {
	mock.Mock
}

// Abs provides a mock function with given fields: path
func (_m *FileSystemManager) Abs(path string) (string, error) {
	ret := _m.Called(path)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(path)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(path)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ReadFile provides a mock function with given fields: path
func (_m *FileSystemManager) ReadFile(path string) ([]byte, error) {
	ret := _m.Called(path)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(string) []byte); ok {
		r0 = rf(path)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(path)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Walk provides a mock function with given fields: root, walkFn
func (_m *FileSystemManager) Walk(root string, walkFn filepath.WalkFunc) error {
	ret := _m.Called(root, walkFn)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, filepath.WalkFunc) error); ok {
		r0 = rf(root, walkFn)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
