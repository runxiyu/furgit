package reachability

import (
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
)

// Reachability provides graph traversal over objects in one object store.
//
// It is not safe for concurrent use.
type Reachability struct {
	store objectstore.Store
}

// New builds a Reachability  over one object store.
func New(store objectstore.Store) *Reachability {
	return &Reachability{store: store}
}

// IsAncestor reports whether ancestor is reachable from descendant via commit
// parent edges.
//
// Both inputs are peeled through annotated tags before commit traversal.
func (r *Reachability) IsAncestor(ancestor, descendant objectid.ObjectID) (bool, error) {
	ancestorCommit, err := r.peelRootToDomain(ancestor, DomainCommits)
	if err != nil {
		return false, err
	}
	descendantCommit, err := r.peelRootToDomain(descendant, DomainCommits)
	if err != nil {
		return false, err
	}
	if ancestorCommit == descendantCommit {
		return true, nil
	}

	walk := r.Walk(DomainCommits, nil, map[objectid.ObjectID]struct{}{descendantCommit: {}})
	for id := range walk.Seq() {
		if id == ancestorCommit {
			return true, nil
		}
	}
	if err := walk.Err(); err != nil {
		return false, err
	}
	return false, nil
}

// CheckConnected verifies that all objects reachable from wants (under the
// selected domain) can be fully traversed without missing-object/type/parse
// errors, excluding subgraphs rooted at haves.
func (r *Reachability) CheckConnected(domain Domain, haves, wants map[objectid.ObjectID]struct{}) error {
	walk := r.Walk(domain, haves, wants)
	for range walk.Seq() {
	}
	return walk.Err()
}

// Walk creates one single-use traversal over the selected domain.
func (r *Reachability) Walk(domain Domain, haves, wants map[objectid.ObjectID]struct{}) *Walk {
	walk := &Walk{
		reachability: r,
		domain:       domain,
		haves:        haves,
		wants:        wants,
	}
	if err := validateDomain(domain); err != nil {
		walk.err = err
	}
	return walk
}
