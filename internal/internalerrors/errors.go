package internalerrors

import "errors"

var (
	// ErrNotValid is used when a resource is not valid.
	ErrNotValid = errors.New("not valid")
	// ErrMissing is used when a resource is missing.
	ErrMissing = errors.New("is missing")
)
