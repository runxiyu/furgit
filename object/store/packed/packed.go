package packed

import (
	"errors"
	"os"
	"sync"

	"lindenii.org/go/furgit/internal/cache/clock"
	"lindenii.org/go/furgit/internal/mru"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
)

// ErrMalformedPackedStore reports that
// a pack or pack index in the store is
// truncated, inconsistent, or otherwise corrupt.
var ErrMalformedPackedStore = errors.New("object/store/packed: malformed packed store")

// Packed reads Git objects from pack/index files
// under an objects/pack root.
//
// Packs appearing after construction are only visible
// after an explicit [Packed.Refresh].
//
// Labels: Close-Caller.
type Packed struct {
	// root is the objects/pack directory
	// used for all pack and index file access.
	root *os.Root

	// objectFormat is the expected object format for lookups.
	objectFormat id.ObjectFormat

	// order contains the packs to probe, MRU-first.
	order *mru.Order[*pack]

	// baseCache caches delta bases consumed during resolution.
	baseCache *clock.Clock[baseKey, cachedBase]

	// refreshMu serializes Refresh.
	// Readers uses none of these.
	refreshMu sync.Mutex

	// byName supports reusing surviving packs across Refresh,
	// and retired holds dropped packs until Close,
	// since concurrent readers may still use them.
	byName map[string]*pack

	retired []*pack
}

var _ store.ObjectReader = (*Packed)(nil)

// New creates a packed-object store rooted at an objects/pack directory,
// performing an initial Refresh.
//
// Labels: Deps-Borrowed, Life-Parent.
func New(root *os.Root, objectFormat id.ObjectFormat) (*Packed, error) {
	if objectFormat.Size() == 0 {
		return nil, id.ErrInvalidObjectFormat
	}

	packed := &Packed{
		root:         root,
		objectFormat: objectFormat,
		order:        mru.New[*pack](mru.Options{Interval: 48}),
		baseCache:    newBaseCache(),
		refreshMu:    sync.Mutex{},
		byName:       nil,
		retired:      nil,
	}

	err := packed.Refresh()
	if err != nil {
		return nil, err
	}

	return packed, nil
}

// Close releases mapped pack/index resources associated with the store.
//
// Labels: MT-Unsafe.
func (packed *Packed) Close() error {
	errs := make([]error, 0, len(packed.byName)+len(packed.retired))

	for _, p := range packed.byName {
		errs = append(errs, p.close())
	}

	for _, p := range packed.retired {
		errs = append(errs, p.close())
	}

	packed.byName = nil
	packed.retired = nil

	packed.baseCache.Clear()

	return errors.Join(errs...)
}
