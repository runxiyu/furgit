package commitquery

import objectid "lindenii.org/go/furgit/object/id"

// MergeBase reports one merge base between left and right, if any.
func (queries *Queries) MergeBase(left, right objectid.ObjectID) (objectid.ObjectID, bool, error) {
	query := queries.acquire()
	defer queries.release(query)

	return query.MergeBase(left, right)
}
