// Package mru provides a concurrent most-recently-used ordering over keys.
//
// It expresses recency only,
// never priority,
// and it never evicts.
// Reads are lock-free over an immutable snapshot,
// so a concurrent reorder never perturbs an in-progress walk.
package mru
