//go:build sha1

package furgit

import (
	"crypto/sha1"
)

const HashSize = sha1.Size

var newHash = sha1.Sum
