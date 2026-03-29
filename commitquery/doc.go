// Package commitquery provides commit ancestry and merge-base queries
// over object storage.
//
// It uses commit-ish object IDs, peeling annotated tags when needed,
// and can use an optional commit-graph reader for performance.
package commitquery
