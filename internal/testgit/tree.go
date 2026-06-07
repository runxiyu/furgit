package testgit

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/typ"
)

// MkTreeEntry is one entry of a tree built by [Repo.MkTree].
type MkTreeEntry struct {
	Mode string
	Type typ.Type
	OID  id.ObjectID
	Name string
}

// MkTree builds a tree object from entries and returns its object ID.
func (repo *Repo) MkTree(tb testing.TB, entries []MkTreeEntry) (id.ObjectID, error) {
	tb.Helper()

	var stdin bytes.Buffer
	for _, entry := range entries {
		fmt.Fprintf(&stdin, "%s %s %s\t%s\n", entry.Mode, entry.Type.Name(), entry.OID.String(), entry.Name)
	}

	stdout, err := repo.run(tb, &stdin, "git", "mktree")
	if err != nil {
		return id.ObjectID{}, fmt.Errorf("mktree: %w", err)
	}

	treeID, err := repo.objectFormat.FromString(strings.TrimSuffix(string(stdout), "\n"))
	if err != nil {
		return id.ObjectID{}, fmt.Errorf("parse git mktree output %q: %w", string(stdout), err)
	}

	return treeID, nil
}
