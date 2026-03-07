package mergebase

import (
	"slices"

	"codeberg.org/lindenii/furgit/internal/commitquery"
	"codeberg.org/lindenii/furgit/internal/peel"
	"codeberg.org/lindenii/furgit/objectid"
)

// All returns all merge bases in Git's merge-base --all order.
func (query *Bases) All() ([]objectid.ObjectID, error) {
	if query.computed {
		return slices.Clone(query.bases), query.err
	}

	query.computed = true

	leftCommit, err := peel.ToCommit(query.store, query.left)
	if err != nil {
		query.err = err

		return nil, err
	}

	rightCommit, err := peel.ToCommit(query.store, query.right)
	if err != nil {
		query.err = err

		return nil, err
	}

	ctx := commitquery.NewContext(query.store, query.graph)

	leftIdx, err := ctx.ResolveOID(leftCommit)
	if err != nil {
		query.err = err

		return nil, err
	}

	rightIdx, err := ctx.ResolveOID(rightCommit)
	if err != nil {
		query.err = err

		return nil, err
	}

	candidates, err := commitquery.MergeBases(ctx, leftIdx, rightIdx)
	if err != nil {
		query.err = err

		return nil, err
	}

	slices.SortFunc(candidates, func(left, right commitquery.NodeIndex) int {
		switch {
		case ctx.CommitTime(left) > ctx.CommitTime(right):
			return -1
		case ctx.CommitTime(left) < ctx.CommitTime(right):
			return 1
		default:
			return objectid.Compare(ctx.ID(left), ctx.ID(right))
		}
	})

	query.bases = make([]objectid.ObjectID, 0, len(candidates))
	for _, idx := range candidates {
		query.bases = append(query.bases, ctx.ID(idx))
	}

	return slices.Clone(query.bases), nil
}
