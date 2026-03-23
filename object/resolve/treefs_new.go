package resolve

import "codeberg.org/lindenii/furgit/objectid"

// TreeFS returns one new filesystem view rooted at root, which may be any
// tree-ish object accepted by PeelToTreeID.
func (r *Resolver) TreeFS(root objectid.ObjectID) (*TreeFS, error) {
	rootTree, err := r.PeelToTreeID(root)
	if err != nil {
		return nil, err
	}

	return &TreeFS{
		resolver: r,
		rootTree: rootTree,
	}, nil
}
