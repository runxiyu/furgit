package service

import "codeberg.org/lindenii/furgit/objectid"

// CommandResult is one per-command execution result.
type CommandResult struct {
	Name         string
	Error        string
	RefName      string
	OldID        *objectid.ObjectID
	NewID        *objectid.ObjectID
	ForcedUpdate bool
}
