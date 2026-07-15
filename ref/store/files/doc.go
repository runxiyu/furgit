// Package files provides one Git files-backend reference store,
// reading loose reference files layered over one packed-refs file,
// and writing through lock-file-coordinated transactions.
//
// A store is rooted at one per-worktree git directory
// plus one common git directory;
// see [New].
//
// # Limitations
//
// The store does not write reflogs,
// although deleting a reference removes its stale reflog file.
// It does not compact loose references into packed-refs.
// It does not adjust file modes for core.sharedRepository,
// and does not detect case-insensitive-filesystem name conflicts.
package files
