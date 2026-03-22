package service

import "io"

// Request is one protocol-independent receive-pack execution request.
//
// If PackExpected is true, Pack must be non-nil and remain valid until
// Execute finishes consuming it.
type Request struct {
	Commands     []Command
	PushOptions  []string
	Atomic       bool
	DeleteOnly   bool
	PackExpected bool
	Pack         io.Reader
}
