package commit

import (
	"bytes"
	"slices"
)

// Clone returns a deep copy of the commit
// whose byte fields are independent of any memory the original may alias.
//
// Labels: Life-Independent.
func (commit *Commit) Clone() *Commit {
	clone := &Commit{
		Tree:      commit.Tree,
		Parents:   slices.Clone(commit.Parents),
		Author:    commit.Author.Clone(),
		Committer: commit.Committer.Clone(),
		Message:   bytes.Clone(commit.Message),
		ChangeID:  bytes.Clone(commit.ChangeID),
	}

	if commit.ExtraHeaders != nil {
		clone.ExtraHeaders = make([]ExtraHeader, len(commit.ExtraHeaders))
		for i, h := range commit.ExtraHeaders {
			clone.ExtraHeaders[i] = ExtraHeader{
				Key:   bytes.Clone(h.Key),
				Value: bytes.Clone(h.Value),
			}
		}
	}

	return clone
}
