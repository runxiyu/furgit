package pktline

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidLength indicates a malformed 4-byte hexadecimal length header.
	ErrInvalidLength = errors.New("pktline: invalid length header")
	// ErrTooLarge indicates a payload exceeds configured packet data limits.
	ErrTooLarge = errors.New("pktline: payload too large")
)

// ProtocolError reports invalid pkt-line framing.
//
// It is returned for protocol violations such as invalid control values
// (for example 0003) or non-hex length headers.
type ProtocolError struct {
	Header [4]byte
	Reason string
}

func (e *ProtocolError) Error() string {
	if e == nil {
		return "<nil>"
	}

	if e.Reason == "" {
		return "pktline: protocol error"
	}

	return fmt.Sprintf("pktline: protocol error: %s", e.Reason)
}
