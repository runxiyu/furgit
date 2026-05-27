package commitquery

import objectid "lindenii.org/go/furgit/object/id"

// MergeBase reports one merge base between left and right, if any.
func (query *query) MergeBase(left, right objectid.ObjectID) (objectid.ObjectID, bool, error) {
	bases, err := query.MergeBases(left, right)
	if err != nil {
		return objectid.ObjectID{}, false, err
	}

	if len(bases) == 0 {
		return objectid.ObjectID{}, false, nil
	}

	return bases[0], true, nil
}
