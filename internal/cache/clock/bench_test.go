package clock //nolint:testpackage

import (
	"sync/atomic"
	"testing"

	"lindenii.org/go/lgo/intconv"
)

func benchWeight(_, _ int) uint64 { return 1 }

// goroutineSeq hands each parallel worker a distinct starting offset
// so workers don't go in lockstep over identical keys.
//
//nolint:gochecknoglobals
var goroutineSeq atomic.Uint64

func workerOffset() int {
	result, err := intconv.Uint64ToInt(goroutineSeq.Add(1) * 0x9E3779B1)
	if err != nil {
		panic(err)
	}

	return result
}

func BenchmarkReadHeavy(b *testing.B) {
	// ≈95% Get hits, ≈5% Add, and fits budget.
	const n = 100_000

	clock := New(n, benchWeight)
	for k := range n {
		clock.Add(k, k)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := workerOffset()
		for pb.Next() {
			i++

			key := i % n
			if i%20 == 0 {
				clock.Add(key, key)
			} else {
				_, _ = clock.Get(key)
			}
		}
	})
}

func BenchmarkHotKey(b *testing.B) {
	// Every worker does Get over a relatively small hot set.
	const (
		hot       = 64
		maxWeight = 4096
	)

	clock := New(maxWeight, benchWeight)
	for k := range hot {
		clock.Add(k, k)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := workerOffset()
		for pb.Next() {
			i++
			_, _ = clock.Get(i % hot)
		}
	})
}

func BenchmarkMixed(b *testing.B) {
	// Even split of Get and Add over a working set that fits.
	const n = 100_000

	clock := New(n, benchWeight)
	for k := range n {
		clock.Add(k, k)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := workerOffset()
		for pb.Next() {
			i++

			key := i % n
			if i%2 == 0 {
				clock.Add(key, key)
			} else {
				_, _ = clock.Get(key)
			}
		}
	})
}

func BenchmarkChurn(b *testing.B) {
	// Every op inserts a fresh key into a small budget.
	// I don't think this is likely to happen for git delta clocks,
	// but this may matter if we use this for other workloads later.
	const maxWeight = 4096

	clock := New(maxWeight, benchWeight)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		base := workerOffset()

		i := 0
		for pb.Next() {
			i++
			clock.Add(base+i, i)
		}
	})
}
