package fetch

import objectid "codeberg.org/lindenii/furgit/object/id"

// TreeFS returns a new filesystem view rooted at root, which may be any
// tree-ish object accepted by PeelToTreeID.
func (r *Fetcher) TreeFS(root objectid.ObjectID) (*TreeFS, error) {
	rootTree, err := r.PeelToTreeID(root)
	if err != nil {
		return nil, err
	}

	return &TreeFS{
		resolver: r,
		rootTree: rootTree,
	}, nil
}
