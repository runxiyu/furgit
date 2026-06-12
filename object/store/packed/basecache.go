package packed

import (
	"lindenii.org/go/furgit/internal/cache/clock"
	"lindenii.org/go/furgit/internal/format/packfile"
)

// baseCacheMaxWeight bounds the delta base cache weight.
const baseCacheMaxWeight = 96 << 20

// baseKey addresses an entry in one pack
// as a delta base cache key.
type baseKey struct {
	pack   *pack
	offset uint64
}

// cachedBase is a cached delta base, i.e.,
// its resolved object entry type and full content.
//
// content is shared with concurrent users;
// it may only be returned to callers as a copy.
//
// Labels: Mut-No.
type cachedBase struct {
	entryType packfile.EntryType
	content   []byte
}

// newBaseCache creates the delta base cache.
func newBaseCache() *clock.Clock[baseKey, cachedBase] {
	return clock.New(baseCacheMaxWeight, baseWeight)
}

// baseWeight weighs one cached delta base.
func baseWeight(_ baseKey, base cachedBase) uint64 {
	return uint64(len(base.content)) + 32
}
