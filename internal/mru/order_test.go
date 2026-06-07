package mru_test

import (
	"slices"
	"sync"
	"testing"

	"lindenii.org/go/furgit/internal/mru"
)

func set(keys ...string) map[string]struct{} {
	present := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		present[key] = struct{}{}
	}

	return present
}

func TestTouchMovesToFront(t *testing.T) {
	t.Parallel()

	order := mru.New[string]()
	order.Sync(set("a", "b", "c"))

	order.Touch("a")

	if got := order.Keys(); len(got) == 0 || got[0] != "a" {
		t.Fatalf("after Touch(a), order = %v, want a first", got)
	}

	order.Touch("b")

	if got := order.Keys(); len(got) < 2 || got[0] != "b" || got[1] != "a" {
		t.Fatalf("after Touch(b), order = %v, want [b a ...]", got)
	}
}

func TestSyncDropsAbsentAndKeepsSurvivorOrder(t *testing.T) {
	t.Parallel()

	order := mru.New[string]()
	order.Sync(set("a", "b", "c"))

	// Establish a deterministic recency order: c, b, a.
	order.Touch("a")
	order.Touch("b")
	order.Touch("c")

	if got, want := order.Keys(), []string{"c", "b", "a"}; !slices.Equal(got, want) {
		t.Fatalf("order = %v, want %v", got, want)
	}

	order.Sync(set("a", "c"))

	if got, want := order.Keys(), []string{"c", "a"}; !slices.Equal(got, want) {
		t.Fatalf("after Sync, order = %v, want %v", got, want)
	}
}

func TestTouchAbsentIsNoOp(t *testing.T) {
	t.Parallel()

	order := mru.New[string]()
	order.Sync(set("a", "b"))
	order.Touch("a")
	order.Touch("z")

	if got := order.Keys(); order.Len() != 2 || got[0] != "a" {
		t.Fatalf("after Touch(z), order = %v, want front a and len 2", got)
	}
}

func TestKeysIsConsistentSnapshot(t *testing.T) {
	t.Parallel()

	order := mru.New[string]()
	order.Sync(set("a", "b"))

	snapshot := order.Keys()
	order.Sync(set("c"))

	got := slices.Clone(snapshot)
	slices.Sort(got)

	if want := []string{"a", "b"}; !slices.Equal(got, want) {
		t.Fatalf("captured snapshot = %v, want %v despite later Sync", got, want)
	}

	if cur, want := order.Keys(), []string{"c"}; !slices.Equal(cur, want) {
		t.Fatalf("current order = %v, want %v", cur, want)
	}
}

func TestConcurrentTouchAndKeys(t *testing.T) {
	t.Parallel()

	order := mru.New[string]()
	order.Sync(set("a", "b", "c", "d"))

	var wg sync.WaitGroup

	for range 8 {
		wg.Go(func() {
			for range 1000 {
				for _, key := range order.Keys() {
					order.Touch(key)
				}
			}
		})
	}

	wg.Wait()

	got := slices.Clone(order.Keys())
	slices.Sort(got)

	if want := []string{"a", "b", "c", "d"}; !slices.Equal(got, want) {
		t.Fatalf("membership corrupted: %v", got)
	}
}
