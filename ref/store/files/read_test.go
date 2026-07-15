package files_test

import (
	"errors"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/ref"
	"lindenii.org/go/furgit/ref/store"
)

func TestResolveLooseDirect(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newFilesRepo(t, objectFormat)
			commitID := makeCommit(t, repo, "c1")

			err := repo.UpdateRef(t, "refs/heads/main", commitID)
			if err != nil {
				t.Fatalf("UpdateRef: %v", err)
			}

			filesStore := openStore(t, repo, objectFormat)

			resolved, err := filesStore.Resolve("refs/heads/main")
			if err != nil {
				t.Fatalf("Resolve: %v", err)
			}

			direct, ok := resolved.(ref.Direct)
			if !ok {
				t.Fatalf("Resolve = %T, want ref.Direct", resolved)
			}

			if direct.ID != commitID {
				t.Fatalf("ID = %v, want %v", direct.ID, commitID)
			}

			if direct.PeelState != ref.PeelUnknown {
				t.Fatalf("PeelState = %d, want PeelUnknown", direct.PeelState)
			}
		})
	}
}

func TestResolveSymbolicHead(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newFilesRepo(t, objectFormat)
			commitID := makeCommit(t, repo, "c1")

			err := repo.UpdateRef(t, "refs/heads/main", commitID)
			if err != nil {
				t.Fatalf("UpdateRef: %v", err)
			}

			err = repo.SymbolicRefUpdate(t, "HEAD", "refs/heads/main")
			if err != nil {
				t.Fatalf("SymbolicRefUpdate: %v", err)
			}

			filesStore := openStore(t, repo, objectFormat)

			resolved, err := filesStore.Resolve("HEAD")
			if err != nil {
				t.Fatalf("Resolve(HEAD): %v", err)
			}

			symbolic, ok := resolved.(ref.Symbolic)
			if !ok {
				t.Fatalf("Resolve(HEAD) = %T, want ref.Symbolic", resolved)
			}

			if symbolic.Target != "refs/heads/main" {
				t.Fatalf("Target = %q, want refs/heads/main", symbolic.Target)
			}

			direct, err := filesStore.ResolveToDirect("HEAD")
			if err != nil {
				t.Fatalf("ResolveToDirect(HEAD): %v", err)
			}

			if direct.ID != commitID {
				t.Fatalf("ResolveToDirect ID = %v, want %v", direct.ID, commitID)
			}
		})
	}
}

func TestResolvePackedOnly(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newFilesRepo(t, objectFormat)
			commitID := makeCommit(t, repo, "c1")

			err := repo.UpdateRef(t, "refs/heads/packed", commitID)
			if err != nil {
				t.Fatalf("UpdateRef: %v", err)
			}

			err = repo.PackRefs(t)
			if err != nil {
				t.Fatalf("PackRefs: %v", err)
			}

			filesStore := openStore(t, repo, objectFormat)

			direct, err := filesStore.ResolveToDirect("refs/heads/packed")
			if err != nil {
				t.Fatalf("ResolveToDirect: %v", err)
			}

			if direct.ID != commitID {
				t.Fatalf("ID = %v, want %v", direct.ID, commitID)
			}

			if direct.PeelState != ref.PeelNone {
				t.Fatalf("PeelState = %d, want PeelNone under fully-peeled", direct.PeelState)
			}
		})
	}
}

func TestResolveLooseShadowsPacked(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newFilesRepo(t, objectFormat)
			packedID := makeCommit(t, repo, "packed")
			looseID := makeCommit(t, repo, "loose")

			err := repo.UpdateRef(t, "refs/heads/shadowed", packedID)
			if err != nil {
				t.Fatalf("UpdateRef(packed): %v", err)
			}

			err = repo.PackRefs(t)
			if err != nil {
				t.Fatalf("PackRefs: %v", err)
			}

			err = repo.UpdateRef(t, "refs/heads/shadowed", looseID)
			if err != nil {
				t.Fatalf("UpdateRef(loose): %v", err)
			}

			filesStore := openStore(t, repo, objectFormat)

			direct, err := filesStore.ResolveToDirect("refs/heads/shadowed")
			if err != nil {
				t.Fatalf("ResolveToDirect: %v", err)
			}

			if direct.ID != looseID {
				t.Fatalf("ID = %v, want loose %v", direct.ID, looseID)
			}
		})
	}
}

func TestResolveMissing(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newFilesRepo(t, objectFormat)
			filesStore := openStore(t, repo, objectFormat)

			_, err := filesStore.Resolve("refs/heads/missing")
			if !errors.Is(err, store.ErrReferenceNotFound) {
				t.Fatalf("Resolve err = %v, want ErrReferenceNotFound", err)
			}

			_, err = filesStore.Resolve("")
			if !errors.Is(err, store.ErrReferenceNotFound) {
				t.Fatalf("Resolve(\"\") err = %v, want ErrReferenceNotFound", err)
			}
		})
	}
}

func TestResolvePackedPeeled(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newFilesRepo(t, objectFormat)
			commitID := makeCommit(t, repo, "c1")

			tagID, err := repo.TagAnnotated(t, "v1", commitID, testgit.TagAnnotatedOptions{Message: "tag v1"})
			if err != nil {
				t.Fatalf("TagAnnotated: %v", err)
			}

			err = repo.PackRefs(t)
			if err != nil {
				t.Fatalf("PackRefs: %v", err)
			}

			filesStore := openStore(t, repo, objectFormat)

			direct, err := filesStore.ResolveToDirect("refs/tags/v1")
			if err != nil {
				t.Fatalf("ResolveToDirect: %v", err)
			}

			if direct.ID != tagID {
				t.Fatalf("ID = %v, want tag object %v", direct.ID, tagID)
			}

			if direct.PeelState != ref.PeelTo {
				t.Fatalf("PeelState = %d, want PeelTo", direct.PeelState)
			}

			if direct.PeeledID != commitID {
				t.Fatalf("PeeledID = %v, want %v", direct.PeeledID, commitID)
			}
		})
	}
}

func TestResolveBrokenLoose(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newFilesRepo(t, objectFormat)
			gitdir := openGitDirRoot(t, repo)

			err := gitdir.WriteFile("refs/heads/broken", []byte("garbage\n"), 0o644)
			if err != nil {
				t.Fatalf("WriteFile: %v", err)
			}

			filesStore := openStore(t, repo, objectFormat)

			_, err = filesStore.Resolve("refs/heads/broken")
			if err == nil {
				t.Fatal("Resolve unexpectedly succeeded")
			}

			if errors.Is(err, store.ErrReferenceNotFound) {
				t.Fatalf("Resolve err = %v, want a broken-reference error", err)
			}
		})
	}
}

func TestResolveTrailingData(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newFilesRepo(t, objectFormat)
			commitID := makeCommit(t, repo, "c1")
			gitdir := openGitDirRoot(t, repo)

			content := commitID.String() + "\t\tbranch 'main' of example.org:repo\n"

			err := gitdir.WriteFile("FETCH_HEAD", []byte(content), 0o644)
			if err != nil {
				t.Fatalf("WriteFile: %v", err)
			}

			filesStore := openStore(t, repo, objectFormat)

			direct, err := filesStore.ResolveToDirect("FETCH_HEAD")
			if err != nil {
				t.Fatalf("ResolveToDirect(FETCH_HEAD): %v", err)
			}

			if direct.ID != commitID {
				t.Fatalf("ID = %v, want %v", direct.ID, commitID)
			}
		})
	}
}

func TestResolveSymlinkRef(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newFilesRepo(t, objectFormat)
			commitID := makeCommit(t, repo, "c1")

			err := repo.UpdateRef(t, "refs/heads/main", commitID)
			if err != nil {
				t.Fatalf("UpdateRef: %v", err)
			}

			gitdir := openGitDirRoot(t, repo)

			err = gitdir.Remove("HEAD")
			if err != nil {
				t.Fatalf("Remove(HEAD): %v", err)
			}

			err = gitdir.Symlink("refs/heads/main", "HEAD")
			if err != nil {
				t.Fatalf("Symlink: %v", err)
			}

			filesStore := openStore(t, repo, objectFormat)

			resolved, err := filesStore.Resolve("HEAD")
			if err != nil {
				t.Fatalf("Resolve(HEAD): %v", err)
			}

			symbolic, ok := resolved.(ref.Symbolic)
			if !ok {
				t.Fatalf("Resolve(HEAD) = %T, want ref.Symbolic", resolved)
			}

			if symbolic.Target != "refs/heads/main" {
				t.Fatalf("Target = %q, want refs/heads/main", symbolic.Target)
			}

			direct, err := filesStore.ResolveToDirect("HEAD")
			if err != nil {
				t.Fatalf("ResolveToDirect(HEAD): %v", err)
			}

			if direct.ID != commitID {
				t.Fatalf("ID = %v, want %v", direct.ID, commitID)
			}
		})
	}
}

func TestResolveEmptyDirShadowsPacked(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newFilesRepo(t, objectFormat)
			commitID := makeCommit(t, repo, "c1")

			err := repo.UpdateRef(t, "refs/heads/dir-shadow", commitID)
			if err != nil {
				t.Fatalf("UpdateRef: %v", err)
			}

			err = repo.PackRefs(t)
			if err != nil {
				t.Fatalf("PackRefs: %v", err)
			}

			gitdir := openGitDirRoot(t, repo)

			err = gitdir.MkdirAll("refs/heads/dir-shadow", 0o755)
			if err != nil {
				t.Fatalf("MkdirAll: %v", err)
			}

			filesStore := openStore(t, repo, objectFormat)

			direct, err := filesStore.ResolveToDirect("refs/heads/dir-shadow")
			if err != nil {
				t.Fatalf("ResolveToDirect: %v", err)
			}

			if direct.ID != commitID {
				t.Fatalf("ID = %v, want %v", direct.ID, commitID)
			}
		})
	}
}

func TestResolveToDirectCycle(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newFilesRepo(t, objectFormat)

			err := repo.SymbolicRefUpdate(t, "refs/sym-a", "refs/sym-b")
			if err != nil {
				t.Fatalf("SymbolicRefUpdate(a): %v", err)
			}

			err = repo.SymbolicRefUpdate(t, "refs/sym-b", "refs/sym-a")
			if err != nil {
				t.Fatalf("SymbolicRefUpdate(b): %v", err)
			}

			filesStore := openStore(t, repo, objectFormat)

			_, err = filesStore.ResolveToDirect("refs/sym-a")
			if !errors.Is(err, store.ErrSymbolicCycle) {
				t.Fatalf("ResolveToDirect err = %v, want ErrSymbolicCycle", err)
			}
		})
	}
}
