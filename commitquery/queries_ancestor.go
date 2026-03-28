package commitquery

import objectid "codeberg.org/lindenii/furgit/object/id"

// IsAncestor reports whether ancestor is reachable from descendant through
// commit parent edges.
//
// Both inputs are peeled through annotated tags before commit traversal.
func (queries *Queries) IsAncestor(ancestor, descendant objectid.ObjectID) (bool, error) {
	query := queries.acquire()
	defer queries.release(query)

	return query.IsAncestor(ancestor, descendant)
}
