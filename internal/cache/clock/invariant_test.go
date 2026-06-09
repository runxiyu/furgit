package clock //nolint:testpackage

import "testing"

// checkShard verifies a shard's structural invariants at a quiescent point.
//
// It must be called with no concurrent operations in flight.
func checkShard[K comparable, V any](t *testing.T, shard *shard[K, V]) {
	t.Helper()

	shard.mu.Lock()
	defer shard.mu.Unlock()

	ringLen := 0

	var ringWeight uint64

	seen := make(map[*entry[K, V]]struct{})

	if shard.hand != nil { //nolint:nestif
		for e := shard.hand; ; e = e.next {
			if e.prev == nil || e.next == nil {
				t.Fatalf("nil ring link at key %v", e.key)
			}

			if e.next.prev != e || e.prev.next != e {
				t.Fatalf("ring links not reciprocal at key %v", e.key)
			}

			if _, dup := seen[e]; dup {
				t.Fatalf("ring revisits a node before returning to the hand")
			}

			seen[e] = struct{}{}
			ringLen++
			ringWeight += e.weight

			if got, ok := shard.items.Load(e.key); !ok || got != e {
				t.Fatalf("ring node %v is not mapped to itself", e.key)
			}

			if e.next == shard.hand {
				break
			}
		}
	}

	if ringLen != shard.count {
		t.Fatalf("ring length %d != count %d", ringLen, shard.count)
	}

	if ringWeight != shard.weight {
		t.Fatalf("ring weight %d != shard weight %d", ringWeight, shard.weight)
	}

	if shard.weight > shard.maxWeight {
		t.Fatalf("weight %d exceeds budget %d", shard.weight, shard.maxWeight)
	}

	if (shard.hand == nil) != (shard.count == 0) {
		t.Fatalf("hand/count disagree: hand=%v count=%d", shard.hand, shard.count)
	}

	mapLen := 0

	shard.items.Range(func(_ K, e *entry[K, V]) bool {
		mapLen++

		if _, ok := seen[e]; !ok {
			t.Fatalf("mapped entry %v missing from ring", e.key)
		}

		return true
	})

	if mapLen != shard.count {
		t.Fatalf("map size %d != count %d", mapLen, shard.count)
	}
}

// checkCache verifies every shard's invariants.
func checkCache[K comparable, V any](t *testing.T, cache *Cache[K, V]) {
	t.Helper()

	for _, shard := range cache.shards {
		checkShard(t, shard)
	}
}
