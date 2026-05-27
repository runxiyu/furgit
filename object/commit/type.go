package commit

import objecttype "lindenii.org/go/furgit/object/type"

// ObjectType returns TypeCommit.
func (commit *Commit) ObjectType() objecttype.Type {
	_ = commit

	return objecttype.TypeCommit
}
