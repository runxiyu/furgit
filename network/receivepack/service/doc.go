// Package service implements the protocol-independent receive-pack service.
//
// A Service borrows the stores, roots, hooks, and I/O endpoints supplied in
// Options. Callers retain ownership of those dependencies and must keep them
// valid for each Execute call that uses them.
package service
