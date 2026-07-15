package files_test

import (
	"errors"
	"path/filepath"
	"testing"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/ref"
	"lindenii.org/go/furgit/ref/store"
	"lindenii.org/go/furgit/ref/store/files"
)

func TestWorktreeStores(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newFilesRepo(t, objectFormat)
			mainID := makeCommit(t, repo, "main")
			hotID := makeCommit(t, repo, "hot")

			err := repo.UpdateRef(t, "refs/heads/main", mainID)
			if err != nil {
				t.Fatalf("UpdateRef(main): %v", err)
			}

			err = repo.SymbolicRefUpdate(t, "HEAD", "refs/heads/main")
			if err != nil {
				t.Fatalf("SymbolicRefUpdate: %v", err)
			}

			err = repo.UpdateRef(t, "refs/heads/hot", hotID)
			if err != nil {
				t.Fatalf("UpdateRef(hot): %v", err)
			}

			err = repo.WorktreeAdd(t, filepath.Join(t.TempDir(), "wt1"), "hot")
			if err != nil {
				t.Fatalf("WorktreeAdd: %v", err)
			}

			common := openGitDirRoot(t, repo)

			worktreeGitDir, err := common.OpenRoot("worktrees/wt1")
			if err != nil {
				t.Fatalf("OpenRoot(worktrees/wt1): %v", err)
			}

			t.Cleanup(func() { _ = worktreeGitDir.Close() })

			mainStore := files.New(common, common, objectFormat, files.Options{})
			worktreeStore := files.New(worktreeGitDir, common, objectFormat, files.Options{})

			// Ambient HEAD differs per store.
			mainHead, err := mainStore.ResolveToDirect("HEAD")
			if err != nil {
				t.Fatalf("main ResolveToDirect(HEAD): %v", err)
			}

			if mainHead.ID != mainID {
				t.Fatalf("main HEAD = %v, want %v", mainHead.ID, mainID)
			}

			worktreeHead, err := worktreeStore.ResolveToDirect("HEAD")
			if err != nil {
				t.Fatalf("worktree ResolveToDirect(HEAD): %v", err)
			}

			if worktreeHead.ID != hotID {
				t.Fatalf("worktree HEAD = %v, want %v", worktreeHead.ID, hotID)
			}

			// Cross-worktree names work from both stores.
			crossHead, err := mainStore.Resolve("worktrees/wt1/HEAD")
			if err != nil {
				t.Fatalf("main Resolve(worktrees/wt1/HEAD): %v", err)
			}

			symbolic, ok := crossHead.(ref.Symbolic)
			if !ok || symbolic.Target != "refs/heads/hot" {
				t.Fatalf("worktrees/wt1/HEAD = %#v, want symbolic to refs/heads/hot", crossHead)
			}

			mainFromWorktree, err := worktreeStore.ResolveToDirect("main-worktree/HEAD")
			if err != nil {
				t.Fatalf("worktree ResolveToDirect(main-worktree/HEAD): %v", err)
			}

			if mainFromWorktree.ID != mainID {
				t.Fatalf("main-worktree/HEAD = %v, want %v", mainFromWorktree.ID, mainID)
			}

			// Shared refs are the same through both stores.
			shared, err := worktreeStore.ResolveToDirect("refs/heads/main")
			if err != nil {
				t.Fatalf("worktree ResolveToDirect(refs/heads/main): %v", err)
			}

			if shared.ID != mainID {
				t.Fatalf("shared ref = %v, want %v", shared.ID, mainID)
			}
		})
	}
}

func TestWorktreePerWorktreeRefs(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newFilesRepo(t, objectFormat)
			mainID := makeCommit(t, repo, "main")
			badID := makeCommit(t, repo, "bad")

			err := repo.UpdateRef(t, "refs/heads/main", mainID)
			if err != nil {
				t.Fatalf("UpdateRef: %v", err)
			}

			err = repo.SymbolicRefUpdate(t, "HEAD", "refs/heads/main")
			if err != nil {
				t.Fatalf("SymbolicRefUpdate: %v", err)
			}

			err = repo.UpdateRef(t, "refs/heads/hot", mainID)
			if err != nil {
				t.Fatalf("UpdateRef(hot): %v", err)
			}

			err = repo.WorktreeAdd(t, filepath.Join(t.TempDir(), "wt1"), "hot")
			if err != nil {
				t.Fatalf("WorktreeAdd: %v", err)
			}

			common := openGitDirRoot(t, repo)

			worktreeGitDir, err := common.OpenRoot("worktrees/wt1")
			if err != nil {
				t.Fatalf("OpenRoot(worktrees/wt1): %v", err)
			}

			t.Cleanup(func() { _ = worktreeGitDir.Close() })

			mainStore := files.New(common, common, objectFormat, files.Options{})
			worktreeStore := files.New(worktreeGitDir, common, objectFormat, files.Options{})

			// A per-worktree ref written through the worktree store
			// is invisible under its ambient name in the main store,
			// but visible under its qualified name.
			tx, err := worktreeStore.BeginTransaction()
			if err != nil {
				t.Fatalf("BeginTransaction: %v", err)
			}

			err = tx.Create("refs/bisect/bad", badID)
			if err != nil {
				t.Fatalf("Create queue: %v", err)
			}

			err = tx.Commit()
			if err != nil {
				t.Fatalf("Commit: %v", err)
			}

			direct, err := worktreeStore.ResolveToDirect("refs/bisect/bad")
			if err != nil {
				t.Fatalf("worktree ResolveToDirect(refs/bisect/bad): %v", err)
			}

			if direct.ID != badID {
				t.Fatalf("refs/bisect/bad = %v, want %v", direct.ID, badID)
			}

			_, err = mainStore.Resolve("refs/bisect/bad")
			if !errors.Is(err, store.ErrReferenceNotFound) {
				t.Fatalf("main Resolve(refs/bisect/bad) = %v, want ErrReferenceNotFound", err)
			}

			qualified, err := mainStore.ResolveToDirect("worktrees/wt1/refs/bisect/bad")
			if err != nil {
				t.Fatalf("main ResolveToDirect(worktrees/wt1/refs/bisect/bad): %v", err)
			}

			if qualified.ID != badID {
				t.Fatalf("qualified = %v, want %v", qualified.ID, badID)
			}
		})
	}
}
