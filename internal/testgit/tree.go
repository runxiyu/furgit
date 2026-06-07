package testgit

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/typ"
)

// TreeEntry is one entry of a tree built by [Repo.MkTree]
// or reported by [Repo.LsTree].
type TreeEntry struct {
	Mode string
	Type typ.Type
	OID  id.ObjectID
	Name string
}

// MkTree builds a tree object from entries and returns its object ID.
func (repo *Repo) MkTree(tb testing.TB, entries []TreeEntry) (id.ObjectID, error) {
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

// LsTree lists the top-level entries of a tree object in Git tree order.
func (repo *Repo) LsTree(tb testing.TB, oid id.ObjectID) ([]TreeEntry, error) {
	tb.Helper()

	stdout, err := repo.run(tb, nil, "git", "ls-tree", "-z", "--full-tree", "--end-of-options", oid.String())
	if err != nil {
		return nil, fmt.Errorf("ls-tree: %w", err)
	}

	var entries []TreeEntry

	for record := range bytes.SplitSeq(stdout, []byte{0}) {
		if len(record) == 0 {
			continue
		}

		meta, name, found := bytes.Cut(record, []byte{'\t'})
		if !found {
			return nil, fmt.Errorf("ls-tree: record %q has no tab separator", record)
		}

		fields := bytes.SplitN(meta, []byte{' '}, 3)
		if len(fields) != 3 {
			return nil, fmt.Errorf("ls-tree: record %q has malformed metadata", record)
		}

		ty, err := typ.Parse(string(fields[1]))
		if err != nil {
			return nil, fmt.Errorf("ls-tree: record %q: parse type: %w", record, err)
		}

		entryOID, err := repo.objectFormat.FromString(string(fields[2]))
		if err != nil {
			return nil, fmt.Errorf("ls-tree: record %q: parse oid: %w", record, err)
		}

		entries = append(entries, TreeEntry{
			Mode: string(fields[0]),
			Type: ty,
			OID:  entryOID,
			Name: string(name),
		})
	}

	return entries, nil
}
