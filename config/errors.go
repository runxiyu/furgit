package config

import (
	"errors"
	"fmt"
)

var (
	// ErrMissing indicates that a looked-up key does not exist.
	ErrMissing = errors.New("config: key not found")

	// ErrValueless indicates that a key exists but carries no value.
	ErrValueless = errors.New("config: key has no value")

	// ErrValueEmpty indicates an empty value where one was required.
	ErrValueEmpty = errors.New("config: empty value")

	// ErrValueRange indicates a value outside the representable range.
	ErrValueRange = errors.New("config: value out of range")

	// ErrValueSyntax indicates a malformed value.
	ErrValueSyntax = errors.New("config: invalid value syntax")
)

// ParseError describes a syntactic error in Git config input.
type ParseError struct {
	// Line is the 1-based input line where the error was detected.
	Line int

	reason string
}

func (err *ParseError) Error() string {
	if err.Line > 0 {
		return fmt.Sprintf("config: parse line %d: %s", err.Line, err.reason)
	}

	return "config: parse: " + err.reason
}
