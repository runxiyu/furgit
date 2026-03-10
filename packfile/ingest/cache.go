package ingest

import (
	"codeberg.org/lindenii/furgit/internal/lru"
	"codeberg.org/lindenii/furgit/objecttype"
)

// deltaBaseCacheKey identifies one resolved base by record index.
type deltaBaseCacheKey struct {
	recordIdx int
}

// deltaBaseCacheValue stores one resolved base object payload.
type deltaBaseCacheValue struct {
	realType objecttype.Type
	content  []byte
}

// deltaBaseCache is a bounded LRU for resolved base payloads.
type deltaBaseCache struct {
	lru *lru.Cache[deltaBaseCacheKey, deltaBaseCacheValue]
}

// newDeltaBaseCache creates one bounded base cache.
func newDeltaBaseCache(maxBytes int64) *deltaBaseCache {
	return &deltaBaseCache{
		lru: lru.New(
			maxBytes,
			func(_ deltaBaseCacheKey, value deltaBaseCacheValue) int64 {
				return int64(len(value.content))
			},
			nil,
		),
	}
}

// get returns one cache entry for recordIdx.
func (cache *deltaBaseCache) get(recordIdx int) (objecttype.Type, []byte, bool) {
	value, ok := cache.lru.Get(deltaBaseCacheKey{recordIdx: recordIdx})
	if !ok {
		return objecttype.TypeInvalid, nil, false
	}

	return value.realType, value.content, true
}

// add stores one cache entry for recordIdx.
func (cache *deltaBaseCache) add(recordIdx int, realType objecttype.Type, content []byte) {
	cache.lru.Add(deltaBaseCacheKey{recordIdx: recordIdx}, deltaBaseCacheValue{
		realType: realType,
		content:  content,
	})
}
