package commit

import (
	"codeberg.org/lindenii/furgit/object/typ"
)

// ObjectType returns TypeCommit.
func (commit *Commit) ObjectType() typ.Type {
	_ = commit

	return typ.TypeCommit
}
