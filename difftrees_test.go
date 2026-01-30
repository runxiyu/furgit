package furgit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiffTreesComplexNestedChanges(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	writeTestFile(t, filepath.Join(workDir, "README.md"), "initial readme\n")
	writeTestFile(t, filepath.Join(workDir, "unchanged.txt"), "leave me as-is\n")
	writeTestFile(t, filepath.Join(workDir, "dir", "file_a.txt"), "alpha v1\n")
	writeTestFile(t, filepath.Join(workDir, "dir", "nested", "file_b.txt"), "beta v1\n")
	writeTestFile(t, filepath.Join(workDir, "dir", "nested", "deeper", "file_c.txt"), "gamma v1\n")
	writeTestFile(t, filepath.Join(workDir, "dir", "nested", "deeper", "old.txt"), "old branch\n")
	writeTestFile(t, filepath.Join(workDir, "treeB", "legacy.txt"), "legacy root\n")
	writeTestFile(t, filepath.Join(workDir, "treeB", "sub", "retired.txt"), "retired\n")

	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "add", ".")
	baseTreeHash := gitCmd(t, repoPath, nil, "--work-tree="+workDir, "write-tree")

	writeTestFile(t, filepath.Join(workDir, "README.md"), "updated readme\n")
	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "rm", "-f", "dir/file_a.txt")
	writeTestFile(t, filepath.Join(workDir, "dir", "nested", "file_b.txt"), "beta v2\n")
	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "rm", "-f", "dir/nested/deeper/old.txt")
	writeTestFile(t, filepath.Join(workDir, "dir", "nested", "deeper", "new.txt"), "new branch entry\n")
	writeTestFile(t, filepath.Join(workDir, "dir", "nested", "deeper", "branch", "info.md"), "branch info\n")
	writeTestFile(t, filepath.Join(workDir, "dir", "nested", "deeper", "branch", "subbranch", "leaf.txt"), "leaf data\n")
	writeTestFile(t, filepath.Join(workDir, "dir", "nested", "deeper", "branch", "subbranch", "deep", "final.txt"), "final artifact\n")
	writeTestFile(t, filepath.Join(workDir, "dir", "newchild.txt"), "brand new sibling\n")
	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "rm", "-r", "-f", "treeB")
	writeTestFile(t, filepath.Join(workDir, "features", "alpha", "README.md"), "alpha docs\n")
	writeTestFile(t, filepath.Join(workDir, "features", "alpha", "beta", "gamma.txt"), "gamma payload\n")
	writeTestFile(t, filepath.Join(workDir, "modules", "v2", "core", "main.go"), "package core\n")
	writeTestFile(t, filepath.Join(workDir, "root_addition.txt"), "root level file\n")

	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "add", ".")
	updatedTreeHash := gitCmd(t, repoPath, nil, "--work-tree="+workDir, "write-tree")

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() {
		_ = repo.Close()
	}()

	baseTree := readStoredTree(t, repo, baseTreeHash)
	updatedTree := readStoredTree(t, repo, updatedTreeHash)

	diffs, err := repo.DiffTrees(baseTree, updatedTree)
	if err != nil {
		t.Fatalf("DiffTrees failed: %v", err)
	}

	expected := map[string]diffExpectation{
		"README.md":                                   {kind: TreeDiffEntryKindModified},
		"dir":                                         {kind: TreeDiffEntryKindModified},
		"dir/file_a.txt":                              {kind: TreeDiffEntryKindDeleted, newNil: true},
		"dir/newchild.txt":                            {kind: TreeDiffEntryKindAdded, oldNil: true},
		"dir/nested":                                  {kind: TreeDiffEntryKindModified},
		"dir/nested/file_b.txt":                       {kind: TreeDiffEntryKindModified},
		"dir/nested/deeper":                           {kind: TreeDiffEntryKindModified},
		"dir/nested/deeper/old.txt":                   {kind: TreeDiffEntryKindDeleted, newNil: true},
		"dir/nested/deeper/new.txt":                   {kind: TreeDiffEntryKindAdded, oldNil: true},
		"dir/nested/deeper/branch":                    {kind: TreeDiffEntryKindAdded, oldNil: true},
		"dir/nested/deeper/branch/info.md":            {kind: TreeDiffEntryKindAdded, oldNil: true},
		"dir/nested/deeper/branch/subbranch":          {kind: TreeDiffEntryKindAdded, oldNil: true},
		"dir/nested/deeper/branch/subbranch/leaf.txt": {kind: TreeDiffEntryKindAdded, oldNil: true},
		"dir/nested/deeper/branch/subbranch/deep":     {kind: TreeDiffEntryKindAdded, oldNil: true},
		"dir/nested/deeper/branch/subbranch/deep/final.txt": {
			kind:   TreeDiffEntryKindAdded,
			oldNil: true,
		},
		"features":                      {kind: TreeDiffEntryKindAdded, oldNil: true},
		"features/alpha":                {kind: TreeDiffEntryKindAdded, oldNil: true},
		"features/alpha/README.md":      {kind: TreeDiffEntryKindAdded, oldNil: true},
		"features/alpha/beta":           {kind: TreeDiffEntryKindAdded, oldNil: true},
		"features/alpha/beta/gamma.txt": {kind: TreeDiffEntryKindAdded, oldNil: true},
		"modules":                       {kind: TreeDiffEntryKindAdded, oldNil: true},
		"modules/v2":                    {kind: TreeDiffEntryKindAdded, oldNil: true},
		"modules/v2/core":               {kind: TreeDiffEntryKindAdded, oldNil: true},
		"modules/v2/core/main.go":       {kind: TreeDiffEntryKindAdded, oldNil: true},
		"root_addition.txt":             {kind: TreeDiffEntryKindAdded, oldNil: true},
		"treeB":                         {kind: TreeDiffEntryKindDeleted, newNil: true},
		"treeB/legacy.txt":              {kind: TreeDiffEntryKindDeleted, newNil: true},
		"treeB/sub":                     {kind: TreeDiffEntryKindDeleted, newNil: true},
		"treeB/sub/retired.txt":         {kind: TreeDiffEntryKindDeleted, newNil: true},
	}

	checkDiffs(t, diffs, expected)
}

func TestDiffTreesDirectoryAddDeleteDeep(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	writeTestFile(t, filepath.Join(workDir, "old_dir", "old.txt"), "stale directory\n")
	writeTestFile(t, filepath.Join(workDir, "old_dir", "sub1", "legacy.txt"), "legacy path\n")
	writeTestFile(t, filepath.Join(workDir, "old_dir", "sub1", "nested", "end.txt"), "legacy end\n")

	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "add", ".")
	originalTreeHash := gitCmd(t, repoPath, nil, "--work-tree="+workDir, "write-tree")

	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "rm", "-r", "-f", "old_dir")
	writeTestFile(t, filepath.Join(workDir, "fresh", "alpha", "beta", "new.txt"), "brand new directory\n")
	writeTestFile(t, filepath.Join(workDir, "fresh", "alpha", "docs", "note.md"), "docs note\n")
	writeTestFile(t, filepath.Join(workDir, "fresh", "alpha", "beta", "gamma", "delta.txt"), "delta payload\n")

	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "add", ".")
	nextTreeHash := gitCmd(t, repoPath, nil, "--work-tree="+workDir, "write-tree")

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() {
		_ = repo.Close()
	}()

	originalTree := readStoredTree(t, repo, originalTreeHash)
	nextTree := readStoredTree(t, repo, nextTreeHash)

	diffs, err := repo.DiffTrees(originalTree, nextTree)
	if err != nil {
		t.Fatalf("DiffTrees failed: %v", err)
	}

	expected := map[string]diffExpectation{
		"fresh":                            {kind: TreeDiffEntryKindAdded, oldNil: true},
		"fresh/alpha":                      {kind: TreeDiffEntryKindAdded, oldNil: true},
		"fresh/alpha/beta":                 {kind: TreeDiffEntryKindAdded, oldNil: true},
		"fresh/alpha/beta/new.txt":         {kind: TreeDiffEntryKindAdded, oldNil: true},
		"fresh/alpha/beta/gamma":           {kind: TreeDiffEntryKindAdded, oldNil: true},
		"fresh/alpha/beta/gamma/delta.txt": {kind: TreeDiffEntryKindAdded, oldNil: true},
		"fresh/alpha/docs":                 {kind: TreeDiffEntryKindAdded, oldNil: true},
		"fresh/alpha/docs/note.md":         {kind: TreeDiffEntryKindAdded, oldNil: true},
		"old_dir":                          {kind: TreeDiffEntryKindDeleted, newNil: true},
		"old_dir/old.txt":                  {kind: TreeDiffEntryKindDeleted, newNil: true},
		"old_dir/sub1":                     {kind: TreeDiffEntryKindDeleted, newNil: true},
		"old_dir/sub1/legacy.txt":          {kind: TreeDiffEntryKindDeleted, newNil: true},
		"old_dir/sub1/nested":              {kind: TreeDiffEntryKindDeleted, newNil: true},
		"old_dir/sub1/nested/end.txt":      {kind: TreeDiffEntryKindDeleted, newNil: true},
	}

	checkDiffs(t, diffs, expected)
}

type diffExpectation struct {
	kind   TreeDiffEntryKind
	oldNil bool
	newNil bool
}

func writeTestFile(t *testing.T, path string, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create directory for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}

func readStoredTree(t *testing.T, repo *Repository, hashStr string) *StoredTree {
	t.Helper()
	hash, err := repo.ParseHash(hashStr)
	if err != nil {
		t.Fatalf("ParseHash failed: %v", err)
	}
	obj, err := repo.ReadObject(hash)
	if err != nil {
		t.Fatalf("ReadObject failed: %v", err)
	}
	tree, ok := obj.(*StoredTree)
	if !ok {
		t.Fatalf("expected *StoredTree, got %T", obj)
	}
	return tree
}

func checkDiffs(t *testing.T, diffs []TreeDiffEntry, expected map[string]diffExpectation) {
	t.Helper()
	got := make(map[string]TreeDiffEntry, len(diffs))
	for _, diff := range diffs {
		key := string(diff.Path)
		if _, exists := got[key]; exists {
			t.Fatalf("duplicate diff entry for %q", key)
		}
		got[key] = diff
	}
	if len(got) != len(expected) {
		t.Fatalf("unexpected diff count: got %d, want %d", len(got), len(expected))
	}

	for path, want := range expected {
		diff, ok := got[path]
		if !ok {
			t.Fatalf("missing diff for %q", path)
		}
		if diff.Kind != want.kind {
			t.Errorf("%s kind: got %v, want %v", path, diff.Kind, want.kind)
		}
		if (diff.Old == nil) != want.oldNil {
			t.Errorf("%s old nil mismatch: got %v, want %v", path, diff.Old == nil, want.oldNil)
		}
		if (diff.New == nil) != want.newNil {
			t.Errorf("%s new nil mismatch: got %v, want %v", path, diff.New == nil, want.newNil)
		}
		if diff.Kind == TreeDiffEntryKindModified && diff.Old != nil && diff.New != nil && diff.Old.ID == diff.New.ID {
			t.Errorf("%s: modified entry should change IDs", path)
		}
	}
}
