package reachability

import (
	"codeberg.org/lindenii/furgit/objectid"
)

// Walk is one single-use iterator traversal.
type Walk struct {
	reachability *Reachability
	domain       Domain
	haves        map[objectid.ObjectID]struct{}
	wants        map[objectid.ObjectID]struct{}
	strict       bool

	seqUsed bool
	err     error
}

// Walk creates one single-use traversal over the selected domain.
//
// In DomainCommits, when a commit-graph reader is attached, parent expansion
// may use commit-graph metadata for speed.
func (r *Reachability) Walk(domain Domain, haves, wants map[objectid.ObjectID]struct{}) *Walk {
	walk := &Walk{
		reachability: r,
		domain:       domain,
		haves:        haves,
		wants:        wants,
	}

	err := validateDomain(domain)
	if err != nil {
		walk.err = err
	}

	return walk
}
