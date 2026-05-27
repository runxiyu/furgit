package commitquery

import objectid "lindenii.org/go/furgit/object/id"

// MergeBases reports all merge bases in Git's merge-base --all order.
//
// Both inputs are peeled through annotated tags before commit traversal.
func (queries *Queries) MergeBases(left, right objectid.ObjectID) ([]objectid.ObjectID, error) {
	query := queries.acquire()
	defer queries.release(query)

	return query.MergeBases(left, right)
}
