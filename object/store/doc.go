// Package objectstore provides interfaces for object storage backends.
//
// Concrete implementations generally inherit the contract documented by the
// interfaces they satisfy. Implementation docs focus on additional guarantees
// and implementation-specific behavior.
//
// There is currently no writing-store interface because different
// object store backends have very different models for writing.
// For example, a loose object store can trivially write single loose
// objects, but writing individual objects to a packfile store would
// be extremely wasteful.
//
// At some time, we will have writing-store interfaces.
package objectstore
