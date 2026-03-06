package packed_test

import (
	"bytes"
	"errors"
	"os"
	"slices"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/ref"
	"codeberg.org/lindenii/furgit/refstore"
	"codeberg.org/lindenii/furgit/refstore/packed"
)

func openPackedRefStoreFromRepo(t *testing.T, testRepo *testgit.TestRepo, algo objectid.Algorithm) *packed.Store {
	t.Helper()

	root := testRepo.OpenGitRoot(t)

	store, err := packed.New(root, algo)
	if err != nil {
		t.Fatalf("packed.New: %v", err)
	}

	return store
}

func openPackedRefStoreFromContent(t *testing.T, content string, algo objectid.Algorithm) (*packed.Store, error) {
	t.Helper()

	dir := t.TempDir()

	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatalf("OpenRoot(temp): %v", err)
	}

	defer func() { _ = root.Close() }()

	err = root.WriteFile("packed-refs", []byte(content), 0o644)
	if err != nil {
		t.Fatalf("WriteFile(packed-refs): %v", err)
	}

	return packed.New(root, algo)
}

func TestPackedResolveAndPeeled(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		_, _, commitID := testRepo.MakeCommit(t, "packed refs commit")
		testRepo.UpdateRef(t, "refs/heads/main", commitID)
		tagID := testRepo.TagAnnotated(t, "v1.0.0", commitID, "annotated tag")
		testRepo.PackRefs(t, "--all", "--prune")

		store := openPackedRefStoreFromRepo(t, testRepo, algo)

		resolvedMain, err := store.Resolve("refs/heads/main")
		if err != nil {
			t.Fatalf("Resolve(main): %v", err)
		}

		mainDet, ok := resolvedMain.(ref.Detached)
		if !ok {
			t.Fatalf("Resolve(main) type = %T, want ref.Detached", resolvedMain)
		}

		if mainDet.ID != commitID {
			t.Fatalf("Resolve(main) id = %s, want %s", mainDet.ID, commitID)
		}

		resolvedTag, err := store.Resolve("refs/tags/v1.0.0")
		if err != nil {
			t.Fatalf("Resolve(tag): %v", err)
		}

		tagDet, ok := resolvedTag.(ref.Detached)
		if !ok {
			t.Fatalf("Resolve(tag) type = %T, want ref.Detached", resolvedTag)
		}

		if tagDet.ID != tagID {
			t.Fatalf("Resolve(tag) id = %s, want %s", tagDet.ID, tagID)
		}

		if tagDet.Peeled == nil {
			t.Fatalf("Resolve(tag) peeled = nil, want commit")
		}

		if *tagDet.Peeled != commitID {
			t.Fatalf("Resolve(tag) peeled = %s, want %s", *tagDet.Peeled, commitID)
		}

		fullTag, err := store.ResolveFully("refs/tags/v1.0.0")
		if err != nil {
			t.Fatalf("ResolveFully(tag): %v", err)
		}

		if fullTag.ID != tagDet.ID {
			t.Fatalf("ResolveFully(tag) id = %s, want %s", fullTag.ID, tagDet.ID)
		}

		_, err = store.Resolve("refs/heads/does-not-exist")
		if !errors.Is(err, refstore.ErrReferenceNotFound) {
			t.Fatalf("Resolve(not-found) error = %v", err)
		}
	})
}

func TestPackedListAndShorten(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		_, _, commitID := testRepo.MakeCommit(t, "packed refs list commit")
		testRepo.UpdateRef(t, "refs/heads/main", commitID)
		testRepo.UpdateRef(t, "refs/tags/main", commitID)
		testRepo.UpdateRef(t, "refs/remotes/origin/main", commitID)
		testRepo.PackRefs(t, "--all", "--prune")

		store := openPackedRefStoreFromRepo(t, testRepo, algo)

		all, err := store.List("")
		if err != nil {
			t.Fatalf("List(all): %v", err)
		}

		allNames := make([]string, 0, len(all))
		for _, entry := range all {
			allNames = append(allNames, entry.Name())
		}

		slices.Sort(allNames)

		wantAll := []string{"refs/heads/main", "refs/remotes/origin/main", "refs/tags/main"}
		if !slices.Equal(allNames, wantAll) {
			t.Fatalf("List(all) names = %v, want %v", allNames, wantAll)
		}

		filtered, err := store.List("refs/heads/*")
		if err != nil {
			t.Fatalf("List(pattern): %v", err)
		}

		if len(filtered) != 1 || filtered[0].Name() != "refs/heads/main" {
			t.Fatalf("List(refs/heads/*) = %v, want refs/heads/main only", filtered)
		}

		short, err := store.Shorten("refs/heads/main")
		if err != nil {
			t.Fatalf("Shorten(main): %v", err)
		}

		if short != "heads/main" {
			t.Fatalf("Shorten(main) = %q, want %q", short, "heads/main")
		}

		_, err = store.Shorten("refs/heads/does-not-exist")
		if !errors.Is(err, refstore.ErrReferenceNotFound) {
			t.Fatalf("Shorten(not-found) error = %v", err)
		}
	})
}

func TestPackedListPatternMatrix(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		_, _, commitID := testRepo.MakeCommit(t, "packed refs pattern matrix")
		testRepo.UpdateRef(t, "refs/heads/main", commitID)
		testRepo.UpdateRef(t, "refs/heads/feature/one", commitID)
		testRepo.UpdateRef(t, "refs/notes/review", commitID)
		testRepo.UpdateRef(t, "refs/tags/v1", commitID)
		testRepo.PackRefs(t, "--all", "--prune")

		store := openPackedRefStoreFromRepo(t, testRepo, algo)

		tests := []struct {
			pattern string
			want    []string
		}{
			{
				pattern: "refs/heads/*",
				want:    []string{"refs/heads/main"},
			},
			{
				pattern: "refs/heads/*/*",
				want:    []string{"refs/heads/feature/one"},
			},
			{
				pattern: "refs/*/feature/one",
				want:    []string{"refs/heads/feature/one"},
			},
			{
				pattern: "refs/heads/feat?re/one",
				want:    []string{"refs/heads/feature/one"},
			},
			{
				pattern: "refs/tags/v[0-9]",
				want:    []string{"refs/tags/v1"},
			},
			{
				pattern: "refs/*/*",
				want:    []string{"refs/heads/main", "refs/notes/review", "refs/tags/v1"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.pattern, func(t *testing.T) {
				got, err := store.List(tt.pattern)
				if err != nil {
					t.Fatalf("List(%q): %v", tt.pattern, err)
				}

				gotNames := refNames(got)
				slices.Sort(gotNames)

				wantNames := append([]string(nil), tt.want...)
				slices.Sort(wantNames)

				if !slices.Equal(gotNames, wantNames) {
					t.Fatalf("List(%q) names = %v, want %v", tt.pattern, gotNames, wantNames)
				}
			})
		}
	})
}

func TestPackedParseErrors(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		cases := []struct {
			name string
			data string
		}{
			{
				name: "peeled without ref",
				data: "^" + stringsOfLen("0", algo.HexLen()) + "\n",
			},
			{
				name: "invalid entry",
				data: "not-a-valid-line\n",
			},
			{
				name: "duplicate ref",
				data: stringsOfLen("0", algo.HexLen()) + " refs/heads/main\n" +
					stringsOfLen("1", algo.HexLen()) + " refs/heads/main\n",
			},
		}

		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				_, err := openPackedRefStoreFromContent(t, tt.data, algo)
				if err == nil {
					t.Fatalf("packed.New expected parse error")
				}
			})
		}
	})
}

func TestPackedNewValidation(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatalf("OpenRoot(temp): %v", err)
	}

	defer func() { _ = root.Close() }()

	_, err = packed.New(root, objectid.AlgorithmUnknown)
	if !errors.Is(err, objectid.ErrInvalidAlgorithm) {
		t.Fatalf("packed.New invalid algorithm error = %v", err)
	}

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		_, err = packed.New(root, algo)
		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("packed.New missing packed-refs error = %v", err)
		}
	})
}

func refNames(refs []ref.Ref) []string {
	names := make([]string, 0, len(refs))
	for _, entry := range refs {
		names = append(names, entry.Name())
	}

	return names
}

func stringsOfLen(ch string, n int) string {
	return string(bytes.Repeat([]byte(ch), n))
}
