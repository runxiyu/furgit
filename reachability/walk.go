package reachability

import (
	objectid "codeberg.org/lindenii/furgit/object/id"
)

// Walk is one single-use iterator traversal.
//
// Labels: MT-Unsafe.
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
//
// Walk retains haves and wants as provided.
//
// Labels: Life-Parent.
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
