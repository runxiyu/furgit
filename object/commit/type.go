package commit

import objecttype "codeberg.org/lindenii/furgit/object/type"

// ObjectType returns TypeCommit.
func (commit *Commit) ObjectType() objecttype.Type {
	_ = commit

	return objecttype.TypeCommit
}
