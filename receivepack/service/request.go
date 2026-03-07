package service

import "io"

// Request is one protocol-independent receive-pack execution request.
type Request struct {
	Commands     []Command
	PushOptions  []string
	Atomic       bool
	DeleteOnly   bool
	PackExpected bool
	Pack         io.Reader
}
