package clock //nolint:testpackage

import (
	"sync"
	"testing"
)

func keyValue(key int) int {
	return key*1000003 + 7
}

func TestConcurrentStress(t *testing.T) {
	t.Parallel()

	const (
		maxWeight = 512
		keys      = 400
		workers   = 8
		rounds    = 5000
	)

	cache := New(maxWeight, func(_ int, _ int) uint64 { return 1 })

	var wg sync.WaitGroup

	for worker := range workers {
		wg.Go(func() {
			for i := range rounds {
				key := (worker*7 + i) % keys

				switch i % 4 {
				case 0, 1:
					cache.Add(key, keyValue(key))
				case 2:
					if got, ok := cache.Get(key); ok && got != keyValue(key) {
						t.Errorf("Get(%d) = %d, want %d", key, got, keyValue(key))
					}
				case 3:
					if got, ok := cache.Peek(key); ok && got != keyValue(key) {
						t.Errorf("Peek(%d) = %d, want %d", key, got, keyValue(key))
					}
				}
			}
		})
	}

	wg.Wait()

	checkCache(t, cache)

	if got := cache.Weight(); got > maxWeight {
		t.Fatalf("weight %d exceeds max %d", got, maxWeight)
	}
}

func TestReadDuringEviction(t *testing.T) {
	t.Parallel()

	const (
		maxWeight = 8
		hot       = 64
		writers   = 2
		readers   = 6
		rounds    = 20000
	)

	cache := New(maxWeight, func(_ int, _ int) uint64 { return 1 })

	var wg sync.WaitGroup

	for range writers {
		wg.Go(func() {
			for i := range rounds {
				key := i % hot
				cache.Add(key, keyValue(key))
			}
		})
	}

	for range readers {
		wg.Go(func() {
			for i := range rounds {
				key := i % hot

				if got, ok := cache.Get(key); ok && got != keyValue(key) {
					t.Errorf("Get(%d) = %d, want %d", key, got, keyValue(key))
				}

				if got, ok := cache.Peek(key); ok && got != keyValue(key) {
					t.Errorf("Peek(%d) = %d, want %d", key, got, keyValue(key))
				}
			}
		})
	}

	wg.Wait()

	checkCache(t, cache)

	if got := cache.Weight(); got > maxWeight {
		t.Fatalf("weight %d exceeds max %d", got, maxWeight)
	}
}
