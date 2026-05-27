package commit

import (
	"lindenii.org/go/furgit/object/typ"
)

// ObjectType returns TypeCommit.
func (commit *Commit) ObjectType() typ.Type {
	_ = commit

	return typ.TypeCommit
}
