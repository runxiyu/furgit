package clock_test

import (
	"fmt"
	"strings"
	"testing"

	"lindenii.org/go/furgit/internal/cache/clock"
	"lindenii.org/go/lgo/intconv"
)

func byteWeight(_ string, value string) uint64 {
	weight, err := intconv.IntToUint64(len(value))
	if err != nil {
		return 0
	}

	return weight
}

func TestCacheAddGetPeek(t *testing.T) {
	t.Parallel()

	cache := clock.New(1<<20, byteWeight)

	if !cache.Add("a", "alpha") {
		t.Fatalf("Add(a) should succeed")
	}

	if got, ok := cache.Get("a"); !ok || got != "alpha" {
		t.Fatalf("Get(a) = (%q, %v), want (alpha, true)", got, ok)
	}

	if got, ok := cache.Peek("a"); !ok || got != "alpha" {
		t.Fatalf("Peek(a) = (%q, %v), want (alpha, true)", got, ok)
	}

	if _, ok := cache.Get("missing"); ok {
		t.Fatalf("Get(missing) should miss")
	}
}

func TestCacheWeightStaysBounded(t *testing.T) {
	t.Parallel()

	const maxWeight = 4096

	cache := clock.New(maxWeight, byteWeight)
	value := strings.Repeat("x", 64)

	for i := range 1000 {
		cache.Add(fmt.Sprintf("key-%d", i), value)
	}

	if got := cache.Weight(); got > maxWeight {
		t.Fatalf("weight = %d, exceeds max %d", got, maxWeight)
	}
}

func TestCacheLenAndClear(t *testing.T) {
	t.Parallel()

	cache := clock.New(1<<20, byteWeight)

	for i := range 10 {
		cache.Add(fmt.Sprintf("key-%d", i), "v")
	}

	if got := cache.Len(); got != 10 {
		t.Fatalf("Len = %d, want 10", got)
	}

	cache.Clear()

	if got := cache.Len(); got != 0 {
		t.Fatalf("Len after Clear = %d, want 0", got)
	}

	if got := cache.Weight(); got != 0 {
		t.Fatalf("Weight after Clear = %d, want 0", got)
	}
}

func TestCacheRejectsOversized(t *testing.T) {
	t.Parallel()

	cache := clock.New(4, byteWeight)

	if cache.Add("a", "xxxxx") {
		t.Fatalf("oversized Add should report false")
	}

	if _, ok := cache.Get("a"); ok {
		t.Fatalf("oversized entry must not be cached")
	}

	if got := cache.Weight(); got != 0 {
		t.Fatalf("weight = %d, want 0", got)
	}
}
