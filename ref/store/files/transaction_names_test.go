package files_test

import (
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/ref"
	refstore "codeberg.org/lindenii/furgit/ref/store"
)

func TestFilesTransactionValidateUpdateNames(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		_, _, commitID := testRepo.MakeCommit(t, "base")

		store := openFilesStore(t, testRepo, algo)

		tests := []struct {
			name    string
			queue   func(refstore.Transaction) error
			wantErr bool
		}{
			{
				name: "create refs/heads/main",
				queue: func(tx refstore.Transaction) error {
					return tx.Create("refs/heads/main", commitID)
				},
			},
			{
				name: "create foo/bar",
				queue: func(tx refstore.Transaction) error {
					return tx.Create("foo/bar", commitID)
				},
			},
			{
				name: "create FETCH_HEAD",
				queue: func(tx refstore.Transaction) error {
					return tx.Create("FETCH_HEAD", commitID)
				},
				wantErr: true,
			},
			{
				name: "create MERGE_HEAD",
				queue: func(tx refstore.Transaction) error {
					return tx.Create("MERGE_HEAD", commitID)
				},
				wantErr: true,
			},
			{
				name: "create bad refname",
				queue: func(tx refstore.Transaction) error {
					return tx.Create("refs/heads/.bad", commitID)
				},
				wantErr: true,
			},
			{
				name: "verify unsafe delete-style name",
				queue: func(tx refstore.Transaction) error {
					return tx.Verify("foo/bar", commitID)
				},
				wantErr: true,
			},
			{
				name: "verify pseudoref-style name",
				queue: func(tx refstore.Transaction) error {
					return tx.Verify("PSEUDOREF", commitID)
				},
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tx, err := store.BeginTransaction()
				if err != nil {
					t.Fatalf("BeginTransaction: %v", err)
				}

				err = tt.queue(tx)
				if (err != nil) != tt.wantErr {
					t.Fatalf("queue err=%v, wantErr=%v", err, tt.wantErr)
				}

				_ = tx.Abort()
			})
		}
	})
}

func TestFilesTransactionSymbolicTargetRules(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		_, _, mainID := testRepo.MakeCommit(t, "main")
		testRepo.UpdateRef(t, "refs/heads/main", mainID)
		testRepo.UpdateRef(t, "ORIG_HEAD", mainID)

		store := openFilesStore(t, testRepo, algo)

		tests := []struct {
			name    string
			queue   func(refstore.Transaction) error
			wantErr bool
		}{
			{
				name: "head requires branch target",
				queue: func(tx refstore.Transaction) error {
					return tx.CreateSymbolic("HEAD", "foo")
				},
				wantErr: true,
			},
			{
				name: "head accepts refs/heads target",
				queue: func(tx refstore.Transaction) error {
					return tx.CreateSymbolic("HEAD", "refs/heads/main")
				},
			},
			{
				name: "non-head allows top-level target",
				queue: func(tx refstore.Transaction) error {
					return tx.CreateSymbolic("refs/heads/top-level", "ORIG_HEAD")
				},
			},
			{
				name: "non-head rejects invalid target",
				queue: func(tx refstore.Transaction) error {
					return tx.CreateSymbolic("refs/heads/invalid", "foo..bar")
				},
				wantErr: true,
			},
			{
				name: "non-head allows worktree target",
				queue: func(tx refstore.Transaction) error {
					return tx.CreateSymbolic("refs/heads/worktree-target", "worktrees/wt1/HEAD")
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tx, err := store.BeginTransaction()
				if err != nil {
					t.Fatalf("BeginTransaction: %v", err)
				}

				err = tt.queue(tx)
				if (err != nil) != tt.wantErr {
					t.Fatalf("queue err=%v, wantErr=%v", err, tt.wantErr)
				}

				_ = tx.Abort()
			})
		}

		tx, err := store.BeginTransaction()
		if err != nil {
			t.Fatalf("BeginTransaction(final symbolic): %v", err)
		}

		err = tx.CreateSymbolic("refs/heads/top-level", "ORIG_HEAD")
		if err != nil {
			t.Fatalf("CreateSymbolic(top-level): %v", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("Commit(CreateSymbolic top-level): %v", err)
		}

		got, err := store.Resolve("refs/heads/top-level")
		if err != nil {
			t.Fatalf("Resolve(top-level): %v", err)
		}

		sym, ok := got.(ref.Symbolic)
		if !ok {
			t.Fatalf("Resolve(top-level) type = %T, want ref.Symbolic", got)
		}

		if sym.Target != "ORIG_HEAD" {
			t.Fatalf("top-level target = %q, want %q", sym.Target, "ORIG_HEAD")
		}
	})
}
