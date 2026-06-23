package ingest

import (
	"lindenii.org/go/furgit/internal/cache/clock"
	"lindenii.org/go/furgit/object/typ"
)

const baseCacheMaxWeight = 96 << 20

type baseCacheKey struct {
	offset int
}

type cachedContent struct {
	objectType typ.Type
	content    []byte
}

func newBaseCache() *clock.Clock[baseCacheKey, cachedContent] {
	return clock.New(baseCacheMaxWeight, baseContentWeight)
}

func baseContentWeight(_ baseCacheKey, base cachedContent) uint64 {
	return uint64(len(base.content)) + 32
}
