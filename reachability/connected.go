package reachability

import objectid "codeberg.org/lindenii/furgit/object/id"

// CheckConnected verifies that all objects reachable from wants (under the
// selected domain) can be fully traversed without missing-object/type/parse
// errors, excluding subgraphs rooted at haves.
//
// Even with commit-graph acceleration available, each visited commit is
// still validated against the object store.
func (r *Reachability) CheckConnected(domain Domain, haves, wants map[objectid.ObjectID]struct{}) error {
	walk := r.Walk(domain, haves, wants)

	walk.strict = true
	for range walk.Seq() {
	}

	return walk.Err()
}
