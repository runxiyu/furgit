package testgit

import (
	"fmt"
	"strings"
	"testing"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/typ"
)

// Seeded lists the objects created by one [Repo.SeedHistory], by type.
// Each object ID appears exactly once.
type Seeded struct {
	Blobs   []id.ObjectID
	Trees   []id.ObjectID
	Commits []id.ObjectID
	Tags    []id.ObjectID
}

// All returns all seeded object IDs.
func (seeded *Seeded) All() []id.ObjectID {
	all := make([]id.ObjectID, 0, len(seeded.Blobs)+len(seeded.Trees)+len(seeded.Commits)+len(seeded.Tags))

	all = append(all, seeded.Blobs...)
	all = append(all, seeded.Trees...)
	all = append(all, seeded.Commits...)
	all = append(all, seeded.Tags...)

	return all
}

// SeedHistory creates one deterministic synthetic history
// of 64 commits on refs/heads/main,
// covering all four object types and various situations,
// such as stable, growing, unique, and empty blobs,
// nested subtrees shared across several commits,
// executable and symlink tree entries,
// and an annotated tag every sixteenth commit.
func (repo *Repo) SeedHistory(tb testing.TB) (Seeded, error) {
	tb.Helper()

	seeder := &historySeeder{
		repo: repo,
		tb:   tb,
		seen: make(map[id.ObjectID]struct{}),
		seeded: Seeded{
			Blobs:   nil,
			Trees:   nil,
			Commits: nil,
			Tags:    nil,
		},
	}

	err := seeder.run()
	if err != nil {
		return Seeded{}, err
	}

	return seeder.seeded, nil
}

type historySeeder struct {
	repo   *Repo
	tb     testing.TB
	seen   map[id.ObjectID]struct{}
	seeded Seeded
}

func (seeder *historySeeder) run() error {
	readme, err := seeder.blob("stable readme\n")
	if err != nil {
		return err
	}

	empty, err := seeder.blob("")
	if err != nil {
		return err
	}

	link, err := seeder.blob("README")
	if err != nil {
		return err
	}

	var parents []id.ObjectID

	for i := range 64 {
		commitID, err := seeder.commit(i, readme, empty, link, parents)
		if err != nil {
			return fmt.Errorf("seed commit %d: %w", i, err)
		}

		seeder.seeded.Commits = append(seeder.seeded.Commits, commitID)
		parents = []id.ObjectID{commitID}

		if (i+1)%16 == 0 {
			err = seeder.tag(fmt.Sprintf("v%d", (i+1)/16), commitID)
			if err != nil {
				return fmt.Errorf("seed tag at commit %d: %w", i, err)
			}
		}
	}

	return seeder.repo.UpdateRef(seeder.tb, "refs/heads/main", parents[0])
}

func (seeder *historySeeder) commit(i int, readme, empty, link id.ObjectID, parents []id.ObjectID) (id.ObjectID, error) {
	var growingContent strings.Builder
	for j := range i + 1 {
		fmt.Fprintf(&growingContent, "entry %d\n", j)
	}

	growing, err := seeder.blob(growingContent.String())
	if err != nil {
		return id.ObjectID{}, err
	}

	tool, err := seeder.blob(fmt.Sprintf("tool %d\n", i))
	if err != nil {
		return id.ObjectID{}, err
	}

	inner, err := seeder.blob(fmt.Sprintf("inner %d\n", i/4))
	if err != nil {
		return id.ObjectID{}, err
	}

	subtree, err := seeder.tree([]TreeEntry{
		{Mode: "100644", Type: typ.Blob, OID: inner, Name: "inner"},
		{Mode: "120000", Type: typ.Blob, OID: link, Name: "link"},
	})
	if err != nil {
		return id.ObjectID{}, err
	}

	root, err := seeder.tree([]TreeEntry{
		{Mode: "100644", Type: typ.Blob, OID: readme, Name: "README"},
		{Mode: "100644", Type: typ.Blob, OID: empty, Name: "empty"},
		{Mode: "100644", Type: typ.Blob, OID: growing, Name: "log"},
		{Mode: "040000", Type: typ.Tree, OID: subtree, Name: "sub"},
		{Mode: "100755", Type: typ.Blob, OID: tool, Name: "tool"},
	})
	if err != nil {
		return id.ObjectID{}, err
	}

	return seeder.repo.CommitTree(
		seeder.tb,
		root,
		CommitTreeOptions{Message: fmt.Sprintf("commit %d", i)},
		parents...,
	)
}

func (seeder *historySeeder) blob(content string) (id.ObjectID, error) {
	oid, err := seeder.repo.HashObject(seeder.tb, typ.Blob, strings.NewReader(content))
	if err != nil {
		return id.ObjectID{}, err
	}

	if seeder.record(oid) {
		seeder.seeded.Blobs = append(seeder.seeded.Blobs, oid)
	}

	return oid, nil
}

func (seeder *historySeeder) tree(entries []TreeEntry) (id.ObjectID, error) {
	oid, err := seeder.repo.MkTree(seeder.tb, entries)
	if err != nil {
		return id.ObjectID{}, err
	}

	if seeder.record(oid) {
		seeder.seeded.Trees = append(seeder.seeded.Trees, oid)
	}

	return oid, nil
}

func (seeder *historySeeder) tag(name string, target id.ObjectID) error {
	oid, err := seeder.repo.TagAnnotated(
		seeder.tb,
		name,
		target,
		TagAnnotatedOptions{Message: "tag " + name},
	)
	if err != nil {
		return err
	}

	seeder.seeded.Tags = append(seeder.seeded.Tags, oid)

	return nil
}

func (seeder *historySeeder) record(oid id.ObjectID) bool {
	if _, ok := seeder.seen[oid]; ok {
		return false
	}

	seeder.seen[oid] = struct{}{}

	return true
}
