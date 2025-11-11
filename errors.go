package furgit

import "errors"

var (
	// ErrInvalidObject indicates malformed serialized data.
	ErrInvalidObject = errors.New("furgit: invalid object encoding")
	// ErrInvalidRef indicates malformed refs.
	ErrInvalidRef = errors.New("furgit: invalid ref")
	// ErrNotFound indicates missing refs/objects.
	ErrNotFound = errors.New("furgit: not found")
)
