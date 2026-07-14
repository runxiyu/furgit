package id

import (
	"crypto/sha1" //#nosec:G505
	"crypto/sha256"
	"hash"
)

type objectFormatDetails struct {
	name string
	size int
	sum  func([]byte) ObjectID
	new  func() hash.Hash
}

func (objectFormat ObjectFormat) details() objectFormatDetails {
	return objectFormatTable[objectFormat]
}

//nolint:gochecknoglobals
var objectFormatTable = [...]objectFormatDetails{
	ObjectFormatUnknown: {}, //exhaustruct:ignore
	ObjectFormatSHA1: {
		name: "sha1",
		size: sha1.Size,
		sum: func(data []byte) ObjectID {
			sum := sha1.Sum(data) //#nosec G401

			var id ObjectID
			copy(id.data[:], sum[:])
			id.objectFormat = ObjectFormatSHA1

			return id
		},
		new: sha1.New,
	},
	ObjectFormatSHA256: {
		name: "sha256",
		size: sha256.Size,
		sum: func(data []byte) ObjectID {
			sum := sha256.Sum256(data)

			var id ObjectID
			copy(id.data[:], sum[:])
			id.objectFormat = ObjectFormatSHA256

			return id
		},
		new: sha256.New,
	},
}

// MaxObjectIDSize MUST be >= the largest supported object format size.
const MaxObjectIDSize = sha256.Size

var (
	//nolint:gochecknoglobals
	objectFormatByName = map[string]ObjectFormat{}
	//nolint:gochecknoglobals
	supportedObjectFormats []ObjectFormat
)

func init() { //nolint:gochecknoinits
	for objectFormat := ObjectFormatUnknown + 1; int(objectFormat) < len(objectFormatTable); objectFormat++ {
		info := &objectFormatTable[objectFormat]
		objectFormatByName[info.name] = objectFormat
		supportedObjectFormats = append(supportedObjectFormats, objectFormat)
	}
}
