package refstore

import "fmt"

// InvalidNameError indicates that one requested reference name is invalid.
type InvalidNameError struct {
	Err error
}

func (err *InvalidNameError) Error() string {
	if err == nil || err.Err == nil {
		return "invalid reference name"
	}

	return fmt.Sprintf("invalid reference name: %v", err.Err)
}

func (err *InvalidNameError) Unwrap() error {
	if err == nil {
		return nil
	}

	return err.Err
}

// InvalidValueError indicates that one requested reference value is invalid.
type InvalidValueError struct {
	Err error
}

func (err *InvalidValueError) Error() string {
	if err == nil || err.Err == nil {
		return "invalid reference value"
	}

	return fmt.Sprintf("invalid reference value: %v", err.Err)
}

func (err *InvalidValueError) Unwrap() error {
	if err == nil {
		return nil
	}

	return err.Err
}

// DuplicateUpdateError indicates that one batch or transaction includes a
// duplicate update target.
type DuplicateUpdateError struct{}

func (err *DuplicateUpdateError) Error() string {
	return "duplicate reference update"
}

// CreateExistsError indicates that one create operation targeted an existing
// reference.
type CreateExistsError struct{}

func (err *CreateExistsError) Error() string {
	return "reference already exists"
}

// IncorrectOldValueError indicates that one operation's expected old value did
// not match the current reference value.
type IncorrectOldValueError struct {
	Actual   string
	Expected string
}

func (err *IncorrectOldValueError) Error() string {
	if err == nil {
		return "incorrect old value provided"
	}

	if err.Actual == "" && err.Expected == "" {
		return "incorrect old value provided"
	}

	return fmt.Sprintf("incorrect old value provided: got %q, expected %q", err.Actual, err.Expected)
}

// ExpectedDetachedError indicates that one operation required a detached
// reference but found a different kind.
type ExpectedDetachedError struct{}

func (err *ExpectedDetachedError) Error() string {
	return "expected detached reference"
}

// ExpectedSymbolicError indicates that one operation required a symbolic
// reference but found a different kind.
type ExpectedSymbolicError struct{}

func (err *ExpectedSymbolicError) Error() string {
	return "expected symbolic reference"
}

// NameConflictError indicates that one reference name conflicts with another
// visible or queued reference name.
type NameConflictError struct {
	Other string
}

func (err *NameConflictError) Error() string {
	if err == nil || err.Other == "" {
		return "reference name conflict"
	}

	return fmt.Sprintf("reference name conflict with %q", err.Other)
}
