package service

import objectid "codeberg.org/lindenii/furgit/object/id"

// CommandResult is one per-command execution result.
type CommandResult struct {
	Name         string
	Error        string
	RefName      string
	OldID        *objectid.ObjectID
	NewID        *objectid.ObjectID
	ForcedUpdate bool
}
