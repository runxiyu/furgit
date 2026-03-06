package reachability

import (
	"codeberg.org/lindenii/furgit/objectid"
)

// Walk is one single-use iterator-style traversal.
type Walk struct {
	reachability *Reachability
	domain       Domain
	haves        map[objectid.ObjectID]struct{}
	wants        map[objectid.ObjectID]struct{}
	strict       bool

	seqUsed bool
	err     error
}
