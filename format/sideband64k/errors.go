package sideband64k

import (
	"errors"
	"fmt"
)

var (
	// ErrTooLarge indicates a payload exceeds configured sideband data limits.
	ErrTooLarge = errors.New("sideband64k: payload too large")
	// ErrInvalidBand indicates a data frame has an invalid sideband designator.
	ErrInvalidBand = errors.New("sideband64k: invalid band designator")
)

// ProtocolError reports invalid side-band-64k framing.
type ProtocolError struct {
	Reason string
}

func (e *ProtocolError) Error() string {
	if e == nil {
		return "<nil>"
	}

	if e.Reason == "" {
		return "sideband64k: protocol error"
	}

	return fmt.Sprintf("sideband64k: protocol error: %s", e.Reason)
}
