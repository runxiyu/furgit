package store

import (
	"errors"
	"fmt"

	"lindenii.org/go/furgit/object/id"
)

// ErrReferenceNotFound indicates that a reference does not exist in a backend.
var ErrReferenceNotFound = errors.New("ref/store: reference not found")

// ErrCreateExists indicates that a create operation
// targeted an already-existing reference.
var ErrCreateExists = errors.New("ref/store: reference already exists")

// ErrDuplicateUpdate indicates that one transaction or batch
// includes a duplicate resolved update target.
var ErrDuplicateUpdate = errors.New("ref/store: duplicate reference update")

// ErrExpectedDirect indicates that an operation required a direct reference
// but found a different kind.
var ErrExpectedDirect = errors.New("ref/store: expected direct reference")

// ErrExpectedSymbolic indicates that an operation required a symbolic reference
// but found a different kind.
var ErrExpectedSymbolic = errors.New("ref/store: expected symbolic reference")

// ErrInvalidValue indicates that a requested reference value is invalid,
// such as an empty symbolic target
// or an object ID whose format does not match the store.
var ErrInvalidValue = errors.New("ref/store: invalid reference value")

// ErrSymbolicCycle indicates that resolving a symbolic reference
// encountered a cycle.
var ErrSymbolicCycle = errors.New("ref/store: symbolic reference cycle")

// NameConflictError indicates that one reference name conflicts with another
// visible or queued reference name.
type NameConflictError struct {
	Other string
}

// Error implements error.
func (err *NameConflictError) Error() string {
	return fmt.Sprintf("ref/store: reference name conflict with %q", err.Other)
}

// WrongOldIDError indicates that a direct operation's expected old object ID
// did not match the current reference value.
type WrongOldIDError struct {
	Actual   id.ObjectID
	Expected id.ObjectID
}

// Error implements error.
func (err *WrongOldIDError) Error() string {
	return fmt.Sprintf(
		"ref/store: incorrect old object id: got %s, expected %s",
		err.Actual, err.Expected,
	)
}

// WrongOldTargetError indicates that a symbolic operation's expected old target
// did not match the current reference target.
type WrongOldTargetError struct {
	Actual   string
	Expected string
}

// Error implements error.
func (err *WrongOldTargetError) Error() string {
	return fmt.Sprintf(
		"ref/store: incorrect old target: got %q, expected %q",
		err.Actual, err.Expected,
	)
}
