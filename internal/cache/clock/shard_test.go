package clock //nolint:testpackage

import "testing"

func TestShardEvictsOldest(t *testing.T) {
	t.Parallel()

	shard := newShard[string, string](8)
	shard.add("a", "a", 4)
	shard.add("b", "b", 4)
	shard.add("c", "c", 4) // overflows, evicts

	if _, ok := shard.peek("a"); ok {
		t.Fatalf("a should have been evicted")
	}

	for _, key := range []string{"b", "c"} {
		if _, ok := shard.peek(key); !ok {
			t.Fatalf("%s should remain present", key)
		}
	}

	if got := shard.loadWeight(); got != 8 {
		t.Fatalf("weight = %d, want 8", got)
	}

	if got := shard.len(); got != 2 {
		t.Fatalf("len = %d, want 2", got)
	}
}

func TestShardGetGivesSecondChance(t *testing.T) {
	t.Parallel()

	shard := newShard[string, string](8)
	shard.add("a", "a", 4)
	shard.add("b", "b", 4)
	shard.add("c", "c", 4) // a evicted; b and c survive

	if _, ok := shard.get("b"); !ok {
		t.Fatalf("get(b) should hit")
	}

	shard.add("d", "d", 4) // b was just touched, so c is evicted instead

	if _, ok := shard.peek("c"); ok {
		t.Fatalf("c should have been evicted after b was touched")
	}

	for _, key := range []string{"b", "d"} {
		if _, ok := shard.peek(key); !ok {
			t.Fatalf("%s should remain present", key)
		}
	}
}

func TestShardPeekGivesNoSecondChance(t *testing.T) {
	t.Parallel()

	shard := newShard[string, string](8)
	shard.add("a", "a", 4)
	shard.add("b", "b", 4)
	shard.add("c", "c", 4) // a is evicted; b and c survive with cleared bits

	if _, ok := shard.peek("b"); !ok {
		t.Fatalf("peek(b) should hit")
	}

	shard.add("d", "d", 4) // peek did not refresh b, so b is evicted

	if _, ok := shard.peek("b"); ok {
		t.Fatalf("b should have been evicted; peek must not grant a second chance")
	}

	for _, key := range []string{"c", "d"} {
		if _, ok := shard.peek(key); !ok {
			t.Fatalf("%s should remain present", key)
		}
	}
}

func TestShardReplaceUpdatesWeight(t *testing.T) {
	t.Parallel()

	shard := newShard[string, string](100)
	shard.add("a", "old", 4)
	shard.add("a", "new", 6)

	if got, ok := shard.peek("a"); !ok || got != "new" {
		t.Fatalf("peek(a) = (%q, %v), want (new, true)", got, ok)
	}

	if got := shard.loadWeight(); got != 6 {
		t.Fatalf("weight = %d, want 6", got)
	}

	if got := shard.len(); got != 1 {
		t.Fatalf("len = %d, want 1", got)
	}
}

func TestShardRejectsOversized(t *testing.T) {
	t.Parallel()

	shard := newShard[string, string](5)
	shard.add("a", "x", 3)

	if shard.add("b", "big", 6) {
		t.Fatalf("oversized add should report false")
	}

	if shard.add("a", "huge", 6) {
		t.Fatalf("oversized replace should report false")
	}

	if got, ok := shard.peek("a"); !ok || got != "x" {
		t.Fatalf("peek(a) = (%q, %v), want (x, true); cache must be unchanged", got, ok)
	}

	if _, ok := shard.peek("b"); ok {
		t.Fatalf("b must not have been admitted")
	}

	if got := shard.loadWeight(); got != 3 {
		t.Fatalf("weight = %d, want 3", got)
	}
}

func TestShardClear(t *testing.T) {
	t.Parallel()

	shard := newShard[string, string](100)
	shard.add("a", "a", 4)
	shard.add("b", "b", 4)

	shard.clear()

	if got := shard.loadWeight(); got != 0 {
		t.Fatalf("weight = %d, want 0", got)
	}

	if got := shard.len(); got != 0 {
		t.Fatalf("len = %d, want 0", got)
	}

	if _, ok := shard.peek("a"); ok {
		t.Fatalf("a should be gone after clear")
	}
}
