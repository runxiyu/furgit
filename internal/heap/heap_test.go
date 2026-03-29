package heap_test

import (
	"slices"
	"testing"

	internalheap "codeberg.org/lindenii/furgit/internal/heap"
)

func TestHeapAscending(t *testing.T) {
	t.Parallel()

	heap := internalheap.New(func(left, right int) bool {
		return left < right
	})

	for _, value := range []int{5, 1, 4, 3, 2} {
		heap.Push(value)
	}

	var got []int
	for !heap.Empty() {
		value, ok := heap.Pop()
		if !ok {
			t.Fatal("Pop should succeed")
		}

		got = append(got, value)
	}

	want := []int{1, 2, 3, 4, 5}
	if !slices.Equal(got, want) {
		t.Fatalf("pop order = %v, want %v", got, want)
	}
}

func TestHeapPeek(t *testing.T) {
	t.Parallel()

	heap := internalheap.New(func(left, right int) bool {
		return left > right
	})

	if _, ok := heap.Peek(); ok {
		t.Fatal("Peek on empty heap should miss")
	}

	heap.Push(1)
	heap.Push(3)
	heap.Push(2)

	value, ok := heap.Peek()
	if !ok {
		t.Fatal("Peek should succeed")
	}

	if value != 3 {
		t.Fatalf("Peek = %d, want 3", value)
	}

	if heap.Len() != 3 {
		t.Fatalf("Len after Peek = %d, want 3", heap.Len())
	}
}
