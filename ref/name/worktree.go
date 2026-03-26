package refname

import "strings"

// WorktreeType classifies one worktree-qualified refname prefix.
type WorktreeType uint8

const (
	// WorktreeShared is one ordinary shared refname.
	WorktreeShared WorktreeType = iota

	// WorktreeCurrent is one current-worktree-only refname like HEAD or refs/worktree/...
	WorktreeCurrent

	// WorktreeMain is one main-worktree-qualified refname like main-worktree/HEAD.
	WorktreeMain

	// WorktreeOther is one other-worktree-qualified refname like worktrees/wt1/HEAD.
	WorktreeOther
)

// IsPerWorktree reports whether name is one per-worktree ref namespace.
func IsPerWorktree(name string) bool {
	return strings.HasPrefix(name, "refs/worktree/") ||
		strings.HasPrefix(name, "refs/bisect/") ||
		strings.HasPrefix(name, "refs/rewritten/")
}

// ParsedWorktreeRef is the result of parsing one worktree-qualified refname.
type ParsedWorktreeRef struct {
	Type         WorktreeType
	WorktreeName string
	BareRefName  string
}

// ParseWorktree parses Git's worktree ref prefixes.
func ParseWorktree(name string) ParsedWorktreeRef {
	if bare, ok := strings.CutPrefix(name, "worktrees/"); ok {
		worktreeName, rest, found := strings.Cut(bare, "/")
		if !found {
			return ParsedWorktreeRef{
				Type:         WorktreeOther,
				WorktreeName: worktreeName,
				BareRefName:  "",
			}
		}

		if isCurrentWorktreeRef(rest) {
			return ParsedWorktreeRef{
				Type:         WorktreeOther,
				WorktreeName: worktreeName,
				BareRefName:  rest,
			}
		}
	}

	if bare, ok := strings.CutPrefix(name, "main-worktree/"); ok && isCurrentWorktreeRef(bare) {
		return ParsedWorktreeRef{
			Type:        WorktreeMain,
			BareRefName: bare,
		}
	}

	if isCurrentWorktreeRef(name) {
		return ParsedWorktreeRef{
			Type:        WorktreeCurrent,
			BareRefName: name,
		}
	}

	return ParsedWorktreeRef{
		Type:        WorktreeShared,
		BareRefName: name,
	}
}
