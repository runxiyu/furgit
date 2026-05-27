package reachability

import objectid "lindenii.org/go/furgit/object/id"

// CheckConnected verifies that all objects reachable from wants (under the
// selected domain) can be fully traversed without missing-object/type/parse
// errors, excluding subgraphs rooted at haves.
//
// With commit-graph acceleration available,
// each visited commit is validated against the object store
// iff strict is set to true.
func (r *Reachability) CheckConnected(domain Domain, haves, wants map[objectid.ObjectID]struct{}, strict bool) error {
	walk := r.Walk(domain, haves, wants)

	walk.strict = strict
	for range walk.Seq() {
	}

	return walk.Err()
}
