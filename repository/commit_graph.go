package repository

import (
	"errors"
	"os"

	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
	objectid "codeberg.org/lindenii/furgit/object/id"
)

func openCommitGraph(root *os.Root, algo objectid.Algorithm) (*commitgraphread.Reader, error) {
	reader, err := commitgraphread.Open(root, algo, commitgraphread.OpenChain)
	if err == nil {
		return reader, nil
	}

	var malformed *commitgraphread.MalformedError
	if errors.As(err, &malformed) &&
		malformed.Path == "info/commit-graphs/commit-graph-chain" &&
		malformed.Reason == "missing commit-graph-chain" {
		reader, err = commitgraphread.Open(root, algo, commitgraphread.OpenSingle)
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil //nolint:nilnil
		}

		return reader, err
	}

	return nil, err
}

// CommitGraph returns the configured commit-graph reader, if available.
//
// Labels: Life-Parent, Close-No.
func (repo *Repository) CommitGraph() *commitgraphread.Reader {
	return repo.commitGraph
}
