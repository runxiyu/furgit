package name

import "strings"

// WorktreeType classifies a worktree-qualified name prefix.
type WorktreeType uint8

const (
	// WorktreeShared is a ordinary shared name.
	WorktreeShared WorktreeType = iota

	// WorktreeCurrent is a current-worktree-only name like HEAD or refs/worktree/...
	WorktreeCurrent

	// WorktreeMain is a main-worktree-qualified name like main-worktree/HEAD.
	WorktreeMain

	// WorktreeOther is a other-worktree-qualified name like worktrees/wt1/HEAD.
	WorktreeOther
)

// IsPerWorktree reports whether name is a per-worktree ref namespace.
func IsPerWorktree(name string) bool {
	return strings.HasPrefix(name, "refs/worktree/") ||
		strings.HasPrefix(name, "refs/bisect/") ||
		strings.HasPrefix(name, "refs/rewritten/")
}

func isCurrentWorktreeRef(name string) bool {
	return IsRootSyntax(name) || IsPerWorktree(name)
}

// ParsedWorktreeRef is the result of parsing a worktree-qualified name.
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
			Type:         WorktreeMain,
			WorktreeName: "",
			BareRefName:  bare,
		}
	}

	if isCurrentWorktreeRef(name) {
		return ParsedWorktreeRef{
			Type:         WorktreeCurrent,
			WorktreeName: "",
			BareRefName:  name,
		}
	}

	return ParsedWorktreeRef{
		Type:         WorktreeShared,
		WorktreeName: "",
		BareRefName:  name,
	}
}
