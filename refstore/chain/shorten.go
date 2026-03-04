package chain

import "codeberg.org/lindenii/furgit/refstore"

// Shorten shortens a full reference name using the chain-visible namespace.
func (chain *Chain) Shorten(name string) (string, error) {
	refs, err := chain.List("")
	if err != nil {
		return "", err
	}

	names := make([]string, 0, len(refs))
	found := false

	for _, entry := range refs {
		if entry == nil {
			continue
		}

		full := entry.Name()

		names = append(names, full)
		if full == name {
			found = true
		}
	}

	if !found {
		return "", refstore.ErrReferenceNotFound
	}

	return refstore.ShortenName(name, names), nil
}
