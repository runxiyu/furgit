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

	clock := clock.New(1<<20, byteWeight)

	if !clock.Add("a", "alpha") {
		t.Fatalf("Add(a) should succeed")
	}

	if got, ok := clock.Get("a"); !ok || got != "alpha" {
		t.Fatalf("Get(a) = (%q, %v), want (alpha, true)", got, ok)
	}

	if got, ok := clock.Peek("a"); !ok || got != "alpha" {
		t.Fatalf("Peek(a) = (%q, %v), want (alpha, true)", got, ok)
	}

	if _, ok := clock.Get("missing"); ok {
		t.Fatalf("Get(missing) should miss")
	}
}

func TestCacheWeightStaysBounded(t *testing.T) {
	t.Parallel()

	const maxWeight = 4096

	clock := clock.New(maxWeight, byteWeight)
	value := strings.Repeat("x", 64)

	for i := range 1000 {
		clock.Add(fmt.Sprintf("key-%d", i), value)
	}

	if got := clock.Weight(); got > maxWeight {
		t.Fatalf("weight = %d, exceeds max %d", got, maxWeight)
	}
}

func TestCacheLenAndClear(t *testing.T) {
	t.Parallel()

	clock := clock.New(1<<20, byteWeight)

	for i := range 10 {
		clock.Add(fmt.Sprintf("key-%d", i), "v")
	}

	if got := clock.Len(); got != 10 {
		t.Fatalf("Len = %d, want 10", got)
	}

	clock.Clear()

	if got := clock.Len(); got != 0 {
		t.Fatalf("Len after Clear = %d, want 0", got)
	}

	if got := clock.Weight(); got != 0 {
		t.Fatalf("Weight after Clear = %d, want 0", got)
	}
}

func TestCacheRejectsOversized(t *testing.T) {
	t.Parallel()

	clock := clock.New(4, byteWeight)

	if clock.Add("a", "xxxxx") {
		t.Fatalf("oversized Add should report false")
	}

	if _, ok := clock.Get("a"); ok {
		t.Fatalf("oversized entry must not be clockd")
	}

	if got := clock.Weight(); got != 0 {
		t.Fatalf("weight = %d, want 0", got)
	}
}
