package reachability

import (
	"errors"
	"iter"

	objectid "codeberg.org/lindenii/furgit/object/id"
)

// Seq returns the traversal sequence. It is single-use.
//
// Labels: Life-Parent.
func (walk *Walk) Seq() iter.Seq[objectid.ObjectID] {
	if walk.seqUsed {
		return func(yield func(objectid.ObjectID) bool) {
			_ = yield

			if walk.err == nil {
				walk.err = errors.New("reachability: walk sequence already consumed")
			}
		}
	}

	walk.seqUsed = true

	return func(yield func(objectid.ObjectID) bool) {
		if walk.err != nil {
			return
		}

		stack := walk.initialStack()

		var err error

		visited := make(map[objectid.ObjectID]struct{}, len(stack))
		for len(stack) > 0 {
			item := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			if _, ok := walk.haves[item.id]; ok {
				continue
			}

			if _, ok := visited[item.id]; ok {
				continue
			}

			visited[item.id] = struct{}{}

			var next []walkItem

			next, err = walk.expand(item)
			if err != nil {
				walk.err = err

				return
			}

			if !yield(item.id) {
				return
			}

			stack = append(stack, next...)
		}
	}
}

// Err returns the terminal error, if any, once Seq has been consumed.
func (walk *Walk) Err() error {
	return walk.err
}
