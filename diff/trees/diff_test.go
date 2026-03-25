package trees_test

import (
	"errors"
	"testing"

	"codeberg.org/lindenii/furgit/diff/trees"
	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/object"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/storer/loose"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

func TestDiffComplexNestedChanges(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		repo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: false})

		writeTestFile(t, repo, "README.md", "initial readme\n")
		writeTestFile(t, repo, "unchanged.txt", "leave me as-is\n")
		writeTestFile(t, repo, "dir/file_a.txt", "alpha v1\n")
		writeTestFile(t, repo, "dir/nested/file_b.txt", "beta v1\n")
		writeTestFile(t, repo, "dir/nested/deeper/file_c.txt", "gamma v1\n")
		writeTestFile(t, repo, "dir/nested/deeper/old.txt", "old branch\n")
		writeTestFile(t, repo, "treeB/legacy.txt", "legacy root\n")
		writeTestFile(t, repo, "treeB/sub/retired.txt", "retired\n")

		repo.Run(t, "add", ".")
		baseTreeID := parseID(t, algo, repo.Run(t, "write-tree"))

		writeTestFile(t, repo, "README.md", "updated readme\n")
		repo.Run(t, "rm", "-f", "dir/file_a.txt")
		writeTestFile(t, repo, "dir/nested/file_b.txt", "beta v2\n")
		repo.Run(t, "rm", "-f", "dir/nested/deeper/old.txt")
		writeTestFile(t, repo, "dir/nested/deeper/new.txt", "new branch entry\n")
		writeTestFile(t, repo, "dir/nested/deeper/branch/info.md", "branch info\n")
		writeTestFile(t, repo, "dir/nested/deeper/branch/subbranch/leaf.txt", "leaf data\n")
		writeTestFile(t, repo, "dir/nested/deeper/branch/subbranch/deep/final.txt", "final artifact\n")
		writeTestFile(t, repo, "dir/newchild.txt", "brand new sibling\n")
		repo.Run(t, "rm", "-r", "-f", "treeB")
		writeTestFile(t, repo, "features/alpha/README.md", "alpha docs\n")
		writeTestFile(t, repo, "features/alpha/beta/gamma.txt", "gamma payload\n")
		writeTestFile(t, repo, "modules/v2/core/main.go", "package core\n")
		writeTestFile(t, repo, "root_addition.txt", "root level file\n")

		repo.Run(t, "add", ".")
		updatedTreeID := parseID(t, algo, repo.Run(t, "write-tree"))

		store := openLooseStore(t, repo, algo)
		readTree := makeReadTree(t, store, algo)
		baseTree := mustReadTree(t, readTree, baseTreeID)
		updatedTree := mustReadTree(t, readTree, updatedTreeID)

		diffs, err := trees.Diff(baseTree, updatedTree, readTree)
		if err != nil {
			t.Fatalf("trees.Diff: %v", err)
		}

		expected := map[string]diffExpectation{
			"README.md":                                   {kind: trees.EntryKindModified},
			"dir":                                         {kind: trees.EntryKindModified},
			"dir/file_a.txt":                              {kind: trees.EntryKindDeleted, newNil: true},
			"dir/newchild.txt":                            {kind: trees.EntryKindAdded, oldNil: true},
			"dir/nested":                                  {kind: trees.EntryKindModified},
			"dir/nested/file_b.txt":                       {kind: trees.EntryKindModified},
			"dir/nested/deeper":                           {kind: trees.EntryKindModified},
			"dir/nested/deeper/old.txt":                   {kind: trees.EntryKindDeleted, newNil: true},
			"dir/nested/deeper/new.txt":                   {kind: trees.EntryKindAdded, oldNil: true},
			"dir/nested/deeper/branch":                    {kind: trees.EntryKindAdded, oldNil: true},
			"dir/nested/deeper/branch/info.md":            {kind: trees.EntryKindAdded, oldNil: true},
			"dir/nested/deeper/branch/subbranch":          {kind: trees.EntryKindAdded, oldNil: true},
			"dir/nested/deeper/branch/subbranch/leaf.txt": {kind: trees.EntryKindAdded, oldNil: true},
			"dir/nested/deeper/branch/subbranch/deep":     {kind: trees.EntryKindAdded, oldNil: true},
			"dir/nested/deeper/branch/subbranch/deep/final.txt": {
				kind:   trees.EntryKindAdded,
				oldNil: true,
			},
			"features":                      {kind: trees.EntryKindAdded, oldNil: true},
			"features/alpha":                {kind: trees.EntryKindAdded, oldNil: true},
			"features/alpha/README.md":      {kind: trees.EntryKindAdded, oldNil: true},
			"features/alpha/beta":           {kind: trees.EntryKindAdded, oldNil: true},
			"features/alpha/beta/gamma.txt": {kind: trees.EntryKindAdded, oldNil: true},
			"modules":                       {kind: trees.EntryKindAdded, oldNil: true},
			"modules/v2":                    {kind: trees.EntryKindAdded, oldNil: true},
			"modules/v2/core":               {kind: trees.EntryKindAdded, oldNil: true},
			"modules/v2/core/main.go":       {kind: trees.EntryKindAdded, oldNil: true},
			"root_addition.txt":             {kind: trees.EntryKindAdded, oldNil: true},
			"treeB":                         {kind: trees.EntryKindDeleted, newNil: true},
			"treeB/legacy.txt":              {kind: trees.EntryKindDeleted, newNil: true},
			"treeB/sub":                     {kind: trees.EntryKindDeleted, newNil: true},
			"treeB/sub/retired.txt":         {kind: trees.EntryKindDeleted, newNil: true},
		}

		checkDiffs(t, diffs, expected)
	})
}

func TestDiffDirectoryAddDeleteDeep(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		repo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: false})

		writeTestFile(t, repo, "old_dir/old.txt", "stale directory\n")
		writeTestFile(t, repo, "old_dir/sub1/legacy.txt", "legacy path\n")
		writeTestFile(t, repo, "old_dir/sub1/nested/end.txt", "legacy end\n")

		repo.Run(t, "add", ".")
		originalTreeID := parseID(t, algo, repo.Run(t, "write-tree"))

		repo.Run(t, "rm", "-r", "-f", "old_dir")
		writeTestFile(t, repo, "fresh/alpha/beta/new.txt", "brand new directory\n")
		writeTestFile(t, repo, "fresh/alpha/docs/note.md", "docs note\n")
		writeTestFile(t, repo, "fresh/alpha/beta/gamma/delta.txt", "delta payload\n")

		repo.Run(t, "add", ".")
		nextTreeID := parseID(t, algo, repo.Run(t, "write-tree"))

		store := openLooseStore(t, repo, algo)
		readTree := makeReadTree(t, store, algo)
		originalTree := mustReadTree(t, readTree, originalTreeID)
		nextTree := mustReadTree(t, readTree, nextTreeID)

		diffs, err := trees.Diff(originalTree, nextTree, readTree)
		if err != nil {
			t.Fatalf("trees.Diff: %v", err)
		}

		expected := map[string]diffExpectation{
			"fresh":                            {kind: trees.EntryKindAdded, oldNil: true},
			"fresh/alpha":                      {kind: trees.EntryKindAdded, oldNil: true},
			"fresh/alpha/beta":                 {kind: trees.EntryKindAdded, oldNil: true},
			"fresh/alpha/beta/new.txt":         {kind: trees.EntryKindAdded, oldNil: true},
			"fresh/alpha/beta/gamma":           {kind: trees.EntryKindAdded, oldNil: true},
			"fresh/alpha/beta/gamma/delta.txt": {kind: trees.EntryKindAdded, oldNil: true},
			"fresh/alpha/docs":                 {kind: trees.EntryKindAdded, oldNil: true},
			"fresh/alpha/docs/note.md":         {kind: trees.EntryKindAdded, oldNil: true},
			"old_dir":                          {kind: trees.EntryKindDeleted, newNil: true},
			"old_dir/old.txt":                  {kind: trees.EntryKindDeleted, newNil: true},
			"old_dir/sub1":                     {kind: trees.EntryKindDeleted, newNil: true},
			"old_dir/sub1/legacy.txt":          {kind: trees.EntryKindDeleted, newNil: true},
			"old_dir/sub1/nested":              {kind: trees.EntryKindDeleted, newNil: true},
			"old_dir/sub1/nested/end.txt":      {kind: trees.EntryKindDeleted, newNil: true},
		}

		checkDiffs(t, diffs, expected)
	})
}

type diffExpectation struct {
	kind   trees.EntryKind
	oldNil bool
	newNil bool
}

func writeTestFile(t *testing.T, repo *testgit.TestRepo, path, data string) {
	t.Helper()

	repo.WriteFileAll(t, path, []byte(data), 0o755, 0o644)
}

func openLooseStore(t *testing.T, repo *testgit.TestRepo, algo objectid.Algorithm) *loose.Store {
	t.Helper()

	root := repo.OpenObjectsRoot(t)

	store, err := loose.New(root, algo)
	if err != nil {
		t.Fatalf("loose.New: %v", err)
	}

	t.Cleanup(func() { _ = store.Close() })

	return store
}

func makeReadTree(t *testing.T, store *loose.Store, algo objectid.Algorithm) func(objectid.ObjectID) (*object.Tree, error) {
	t.Helper()

	return func(id objectid.ObjectID) (*object.Tree, error) {
		ty, content, err := store.ReadBytesContent(id)
		if err != nil {
			return nil, err
		}

		if ty != objecttype.TypeTree {
			return nil, errors.New("diff/trees test: object is not a tree")
		}

		return object.ParseTree(content, algo)
	}
}

func mustReadTree(t *testing.T, readTree func(objectid.ObjectID) (*object.Tree, error), id objectid.ObjectID) *object.Tree {
	t.Helper()

	tree, err := readTree(id)
	if err != nil {
		t.Fatalf("read tree %s: %v", id, err)
	}

	return tree
}

func parseID(t *testing.T, algo objectid.Algorithm, hex string) objectid.ObjectID {
	t.Helper()

	id, err := objectid.ParseHex(algo, hex)
	if err != nil {
		t.Fatalf("parse object id %q: %v", hex, err)
	}

	return id
}

func checkDiffs(t *testing.T, diffs []trees.Entry, expected map[string]diffExpectation) {
	t.Helper()

	got := make(map[string]trees.Entry, len(diffs))
	for _, diff := range diffs {
		path := string(diff.Path)
		if _, exists := got[path]; exists {
			t.Fatalf("duplicate diff path %q", path)
		}

		got[path] = diff
	}

	if len(got) != len(expected) {
		t.Fatalf("diff count = %d, want %d", len(got), len(expected))
	}

	for path, want := range expected {
		diff, ok := got[path]
		if !ok {
			t.Fatalf("missing diff for %q", path)
		}

		if diff.Kind != want.kind {
			t.Errorf("%s kind = %v, want %v", path, diff.Kind, want.kind)
		}

		if (diff.Old == nil) != want.oldNil {
			t.Errorf("%s old nil = %v, want %v", path, diff.Old == nil, want.oldNil)
		}

		if (diff.New == nil) != want.newNil {
			t.Errorf("%s new nil = %v, want %v", path, diff.New == nil, want.newNil)
		}

		if diff.Kind == trees.EntryKindModified && diff.Old != nil && diff.New != nil && diff.Old.ID == diff.New.ID {
			t.Errorf("%s modified entry should change IDs", path)
		}
	}
}
