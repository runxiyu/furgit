package service

import "codeberg.org/lindenii/furgit/objectid"

// PlannedUpdate is one ref update that would be applied once ref writing
// exists.
type PlannedUpdate struct {
	Name   string
	OldID  objectid.ObjectID
	NewID  objectid.ObjectID
	Delete bool
}
