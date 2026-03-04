package packed

import "codeberg.org/lindenii/furgit/refstore"

// Shorten returns the shortest unambiguous shorthand for a packed ref name.
func (store *Store) Shorten(name string) (string, error) {
	_, ok := store.byName[name]
	if !ok {
		return "", refstore.ErrReferenceNotFound
	}

	names := make([]string, 0, len(store.ordered))
	for _, entry := range store.ordered {
		names = append(names, entry.Name())
	}

	return refstore.ShortenName(name, names), nil
}
