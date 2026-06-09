// Package clock provides a concurrent, weight-bounded object cache.
//
// The cache is sharded by key,
// and each shard owns an independent fraction of the total budget.
// An entry heavier than that per-shard fraction is never admitted,
// so callers should keep the total budget well above the largest value.
//
// Labels: MT-Safe.
package clock
