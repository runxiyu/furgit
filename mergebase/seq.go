package mergebase

import (
	"errors"
	"iter"

	"codeberg.org/lindenii/furgit/objectid"
)

// Seq returns the merge-base sequence. It is single-use.
func (query *Bases) Seq() iter.Seq[objectid.ObjectID] {
	if query.seqUsed {
		return func(yield func(objectid.ObjectID) bool) {
			_ = yield

			if query.err == nil {
				query.err = errors.New("mergebase: sequence already consumed")
			}
		}
	}

	query.seqUsed = true

	return func(yield func(objectid.ObjectID) bool) {
		if query.err != nil {
			return
		}

		bases, err := query.compute()
		if err != nil {
			query.err = err

			return
		}

		for _, id := range bases {
			if !yield(id) {
				return
			}
		}
	}
}

// Err returns the terminal error, if any, once Seq has been consumed.
func (query *Bases) Err() error {
	return query.err
}
