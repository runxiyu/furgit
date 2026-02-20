package packed

import (
	"codeberg.org/lindenii/furgit/internal/cache/lru"
	"codeberg.org/lindenii/furgit/objecttype"
)

// deltaBaseKey identifies one base object by pack location.
type deltaBaseKey struct {
	packName string
	offset   uint64
}

// deltaBaseValue stores one cached base object body.
type deltaBaseValue struct {
	ty      objecttype.Type
	content []byte
}

// deltaCache wraps a weighted LRU for resolved delta bases.
type deltaCache struct {
	lru *lru.Cache[deltaBaseKey, deltaBaseValue]
}

// newDeltaCache creates a delta base cache with a byte budget.
func newDeltaCache(maxBytes int64) *deltaCache {
	return &deltaCache{
		lru: lru.New(
			maxBytes,
			func(_ deltaBaseKey, value deltaBaseValue) int64 {
				return int64(len(value.content))
			},
			nil,
		),
	}
}

// get returns a cloned cached base object value.
func (cache *deltaCache) get(key deltaBaseKey) (objecttype.Type, []byte, bool) {
	value, ok := cache.lru.Get(key)
	if !ok {
		return objecttype.TypeInvalid, nil, false
	}
	return value.ty, append([]byte(nil), value.content...), true
}

// add stores a cloned base object value.
func (cache *deltaCache) add(key deltaBaseKey, ty objecttype.Type, content []byte) {
	cache.lru.Add(key, deltaBaseValue{
		ty:      ty,
		content: append([]byte(nil), content...),
	})
}

// clear removes all cached entries.
func (cache *deltaCache) clear() {
	cache.lru.Clear()
}
