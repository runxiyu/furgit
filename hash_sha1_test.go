//go:build sha1

package furgit

import (
	"crypto/sha1"
)

const testHashSize = sha1.Size

type (
	testHashType   = [sha1.Size]byte
	TestHash       = Hash[testHashType]
	TestRepository = Repository[testHashType]
	TestBlob       = Blob[testHashType]
	TestTree       = Tree[testHashType]
	TestTreeEntry  = TreeEntry[testHashType]
	TestCommit     = Commit[testHashType]
	TestTag        = Tag[testHashType]
	TestObject     = Object[testHashType]
)
