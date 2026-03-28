package fetch

import (
	"io/fs"

	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/tree"
)

// TreeFS exposes one Git tree as an fs.FS view backed by a Fetcher.
//
// TreeFS interprets names using io/fs path rules. Those rules do not match raw
// Git tree entry naming exactly: names are UTF-8, slash-separated, and must be
// valid fs.FS paths. Tree entries that cannot be represented under those rules
// are not addressable through this API.
//
// Labels: MT-Safe.
type TreeFS struct {
	fetcher   *Fetcher
	rootTree  objectid.ObjectID
	rootEntry *tree.TreeEntry
}

var (
	_ fs.FS         = (*TreeFS)(nil)
	_ fs.ReadFileFS = (*TreeFS)(nil)
	_ fs.ReadDirFS  = (*TreeFS)(nil)
	_ fs.StatFS     = (*TreeFS)(nil)
	_ fs.SubFS      = (*TreeFS)(nil)
)
