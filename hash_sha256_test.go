//go:build !sha1

package furgit

import (
	"crypto/sha256"
)

const testHashSize = sha256.Size

type (
	testHashType   = [sha256.Size]byte
	TestHash       = Hash[testHashType]
	TestRepository = Repository[testHashType]
	TestBlob       = Blob[testHashType]
	TestTree       = Tree[testHashType]
	TestTreeEntry  = TreeEntry[testHashType]
	TestCommit     = Commit[testHashType]
	TestTag        = Tag[testHashType]
	TestObject     = Object[testHashType]
)
