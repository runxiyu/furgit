package service

import objectid "lindenii.org/go/furgit/object/id"

// PlannedUpdate is one requested ref update planned for this execution.
type PlannedUpdate struct {
	Name   string
	OldID  objectid.ObjectID
	NewID  objectid.ObjectID
	Delete bool
}
