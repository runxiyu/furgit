package priorityqueue_test

import (
	"slices"
	"testing"

	"codeberg.org/lindenii/furgit/internal/priorityqueue"
)

func TestQueueAscending(t *testing.T) {
	t.Parallel()

	queue := priorityqueue.New(func(left, right int) bool {
		return left < right
	})

	for _, value := range []int{5, 1, 4, 3, 2} {
		queue.Push(value)
	}

	var got []int
	for queue.Len() > 0 {
		value, ok := queue.Pop()
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
