// Package packed provides a packed refs backend.
package packed

import (
	"codeberg.org/lindenii/furgit/ref"
	"codeberg.org/lindenii/furgit/refstore"
)

// Store reads references from a parsed packed-refs snapshot.
type Store struct {
	byName  map[string]ref.Detached
	ordered []ref.Detached
}

var _ refstore.ReadingStore = (*Store)(nil)
