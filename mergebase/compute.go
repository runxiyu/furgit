package mergebase

import (
	"slices"

	"codeberg.org/lindenii/furgit/internal/commitquery"
	"codeberg.org/lindenii/furgit/internal/peel"
	"codeberg.org/lindenii/furgit/objectid"
)

func (query *Bases) compute() ([]objectid.ObjectID, error) {
	leftCommit, err := peel.ToCommit(query.store, query.left)
	if err != nil {
		return nil, err
	}

	rightCommit, err := peel.ToCommit(query.store, query.right)
	if err != nil {
		return nil, err
	}

	ctx := commitquery.NewContext(query.store, query.graph)

	leftIdx, err := ctx.ResolveOID(leftCommit)
	if err != nil {
		return nil, err
	}

	rightIdx, err := ctx.ResolveOID(rightCommit)
	if err != nil {
		return nil, err
	}

	candidates, err := commitquery.MergeBases(ctx, leftIdx, rightIdx)
	if err != nil {
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

	out := make([]objectid.ObjectID, 0, len(candidates))
	for _, idx := range candidates {
		out = append(out, ctx.ID(idx))
	}

	return out, nil
}
