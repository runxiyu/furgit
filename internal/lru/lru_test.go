package lru_test

import (
	"slices"
	"testing"

	"codeberg.org/lindenii/furgit/internal/lru"
)

type testValue struct {
	weight int64
	label  string
}

func weightFn(key string, value testValue) int64 {
	return value.weight
}

func TestCacheEvictsLRUAndGetUpdatesRecency(t *testing.T) {
	t.Parallel()

	cache := lru.New[string, testValue](8, weightFn, nil)
	cache.Add("a", testValue{weight: 4, label: "a"})
	cache.Add("b", testValue{weight: 4, label: "b"})
	cache.Add("c", testValue{weight: 4, label: "c"})

	if _, ok := cache.Peek("a"); ok {
		t.Fatalf("expected a to be evicted")
	}
	if _, ok := cache.Peek("b"); !ok {
		t.Fatalf("expected b to be present")
	}
	if _, ok := cache.Peek("c"); !ok {
		t.Fatalf("expected c to be present")
	}

	if _, ok := cache.Get("b"); !ok {
		t.Fatalf("Get(b) should hit")
	}
	cache.Add("d", testValue{weight: 4, label: "d"})

	if _, ok := cache.Peek("c"); ok {
		t.Fatalf("expected c to be evicted after b was touched")
	}
	if _, ok := cache.Peek("b"); !ok {
		t.Fatalf("expected b to remain present")
	}
	if _, ok := cache.Peek("d"); !ok {
		t.Fatalf("expected d to be present")
	}
}

func TestCachePeekDoesNotUpdateRecency(t *testing.T) {
	t.Parallel()

	cache := lru.New[string, testValue](4, weightFn, nil)
	cache.Add("a", testValue{weight: 2, label: "a"})
	cache.Add("b", testValue{weight: 2, label: "b"})

	if _, ok := cache.Peek("a"); !ok {
		t.Fatalf("Peek(a) should hit")
	}
	cache.Add("c", testValue{weight: 2, label: "c"})

	if _, ok := cache.Peek("a"); ok {
		t.Fatalf("expected a to be evicted; Peek must not update recency")
	}
	if _, ok := cache.Peek("b"); !ok {
		t.Fatalf("expected b to remain present")
	}
}

func TestCacheReplaceAndResize(t *testing.T) {
	t.Parallel()

	var evicted []string
	cache := lru.New[string, testValue](10, weightFn, func(key string, value testValue) {
		evicted = append(evicted, key+":"+value.label)
	})

	cache.Add("a", testValue{weight: 4, label: "old"})
	cache.Add("b", testValue{weight: 4, label: "b"})
	cache.Add("a", testValue{weight: 6, label: "new"})

	if cache.Weight() != 10 {
		t.Fatalf("Weight() = %d, want 10", cache.Weight())
	}
	if got, ok := cache.Peek("a"); !ok || got.label != "new" {
		t.Fatalf("Peek(a) = (%+v,%v), want new,true", got, ok)
	}
	if !slices.Equal(evicted, []string{"a:old"}) {
		t.Fatalf("evicted = %v, want [a:old]", evicted)
	}

	cache.SetMaxWeight(8)
	if _, ok := cache.Peek("b"); ok {
		t.Fatalf("expected b to be evicted after shrinking max weight")
	}
	if !slices.Equal(evicted, []string{"a:old", "b:b"}) {
		t.Fatalf("evicted = %v, want [a:old b:b]", evicted)
	}
}

func TestCacheRejectsOversizedWithoutMutation(t *testing.T) {
	t.Parallel()

	var evicted []string
	cache := lru.New[string, testValue](5, weightFn, func(key string, value testValue) {
		evicted = append(evicted, key)
	})
	cache.Add("a", testValue{weight: 3, label: "a"})

	if ok := cache.Add("b", testValue{weight: 6, label: "b"}); ok {
		t.Fatalf("Add oversized should return false")
	}
	if got, ok := cache.Peek("a"); !ok || got.label != "a" {
		t.Fatalf("cache should remain unchanged after oversized add")
	}
	if cache.Weight() != 3 {
		t.Fatalf("Weight() = %d, want 3", cache.Weight())
	}
	if len(evicted) != 0 {
		t.Fatalf("evicted = %v, want none", evicted)
	}

	if ok := cache.Add("a", testValue{weight: 6, label: "new"}); ok {
		t.Fatalf("oversized replace should return false")
	}
	if got, ok := cache.Peek("a"); !ok || got.label != "a" {
		t.Fatalf("existing key should remain unchanged after oversized replace")
	}
	if len(evicted) != 0 {
		t.Fatalf("evicted = %v, want none", evicted)
	}
}

func TestCacheRemoveAndClear(t *testing.T) {
	t.Parallel()

	var evicted []string
	cache := lru.New[string, testValue](10, weightFn, func(key string, value testValue) {
		evicted = append(evicted, key)
	})

	cache.Add("a", testValue{weight: 2, label: "a"})
	cache.Add("b", testValue{weight: 3, label: "b"})
	cache.Add("c", testValue{weight: 4, label: "c"})

	removed, ok := cache.Remove("b")
	if !ok || removed.label != "b" {
		t.Fatalf("Remove(b) = (%+v,%v), want b,true", removed, ok)
	}
	if cache.Len() != 2 || cache.Weight() != 6 {
		t.Fatalf("post-remove Len/Weight = %d/%d, want 2/6", cache.Len(), cache.Weight())
	}

	cache.Clear()
	if cache.Len() != 0 || cache.Weight() != 0 {
		t.Fatalf("post-clear Len/Weight = %d/%d, want 0/0", cache.Len(), cache.Weight())
	}

	// Remove emits b, then Clear emits oldest-to-newest among remaining: a, c.
	if !slices.Equal(evicted, []string{"b", "a", "c"}) {
		t.Fatalf("evicted = %v, want [b a c]", evicted)
	}
}

func TestCachePanicsForInvalidConfiguration(t *testing.T) {
	t.Parallel()

	t.Run("negative max", func(t *testing.T) {
		t.Parallel()
		defer func() {
			if recover() == nil {
				t.Fatalf("expected panic")
			}
		}()
		_ = lru.New[string, testValue](-1, weightFn, nil)
	})

	t.Run("nil weight function", func(t *testing.T) {
		t.Parallel()
		defer func() {
			if recover() == nil {
				t.Fatalf("expected panic")
			}
		}()
		_ = lru.New[string, testValue](1, nil, nil)
	})

	t.Run("negative entry weight", func(t *testing.T) {
		t.Parallel()
		cache := lru.New[string, testValue](10, func(_ string, _ testValue) int64 {
			return -1
		}, nil)
		defer func() {
			if recover() == nil {
				t.Fatalf("expected panic")
			}
		}()
		cache.Add("x", testValue{weight: 1, label: "x"})
	})

	t.Run("set negative max", func(t *testing.T) {
		t.Parallel()
		cache := lru.New[string, testValue](10, weightFn, nil)
		defer func() {
			if recover() == nil {
				t.Fatalf("expected panic")
			}
		}()
		cache.SetMaxWeight(-1)
	})
}
