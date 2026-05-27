package files

import (
	"path"

	"lindenii.org/go/furgit/ref/name"
)

func (store *Store) loosePath(name string) refPath {
	parsed := refname.ParseWorktree(name)
	switch parsed.Type {
	case refname.WorktreeCurrent:
		return refPath{root: rootGit, path: parsed.BareRefName}
	case refname.WorktreeMain, refname.WorktreeShared:
		return refPath{root: rootCommon, path: parsed.BareRefName}
	case refname.WorktreeOther:
		return refPath{
			root: rootCommon,
			path: path.Join("worktrees", parsed.WorktreeName, parsed.BareRefName),
		}
	default:
		return refPath{root: rootCommon, path: name}
	}
}
