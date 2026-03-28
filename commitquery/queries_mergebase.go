package commitquery

import objectid "codeberg.org/lindenii/furgit/object/id"

// MergeBases reports all merge bases in Git's merge-base --all order.
//
// Both inputs are peeled through annotated tags before commit traversal.
func (queries *Queries) MergeBases(left, right objectid.ObjectID) ([]objectid.ObjectID, error) {
	query := queries.acquire()
	defer queries.release(query)

	return query.MergeBases(left, right)
}

// MergeBase reports one merge base between left and right, if any.
func (queries *Queries) MergeBase(left, right objectid.ObjectID) (objectid.ObjectID, bool, error) {
	query := queries.acquire()
	defer queries.release(query)

	return query.MergeBase(left, right)
}
