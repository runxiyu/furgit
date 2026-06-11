package fetch

import (
	"io"

	"lindenii.org/go/furgit/errs"
	"lindenii.org/go/furgit/object/blob"
	oid "lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/stored"
	"lindenii.org/go/furgit/object/tag"
	"lindenii.org/go/furgit/object/typ"
)

// ExactBlob reads, parses, and wraps the blob at id.
//
// Labels: Life-Parent.
func (fetcher *Fetcher) ExactBlob(id oid.ObjectID) (*stored.Stored[*blob.Blob], error) {
	parsed, err := fetcher.parseObject(id)
	if err != nil {
		return nil, err
	}

	blob, ok := parsed.(*blob.Blob)
	if !ok {
		return nil, &errs.ObjectTypeError{OID: id, Got: parsed.ObjectType(), Want: typ.Blob}
	}

	return stored.New(id, blob), nil
}

// ExactBlobReader returns a reader for the content of the blob at id,
// together with its content size in bytes.
//
// Labels: Life-Parent, Close-Caller.
func (fetcher *Fetcher) ExactBlobReader(id oid.ObjectID) (io.ReadCloser, uint64, error) {
	return fetcher.exactReader(id, typ.Blob)
}

// PeelToBlob peels tags until it reaches a blob.
//
// Labels: Life-Parent.
func (fetcher *Fetcher) PeelToBlob(id oid.ObjectID) (*stored.Stored[*blob.Blob], error) {
	for {
		obj, err := fetcher.ExactObject(id)
		if err != nil {
			return nil, err
		}

		switch parsed := obj.Object().(type) {
		case *blob.Blob:
			return stored.New(id, parsed), nil
		case *tag.Tag:
			id = parsed.TargetID
		default:
			return nil, &errs.ObjectTypeError{OID: id, Got: parsed.ObjectType(), Want: typ.Blob}
		}
	}
}

// PeelToBlobID peels tags until it reaches a blob object ID.
func (fetcher *Fetcher) PeelToBlobID(id oid.ObjectID) (oid.ObjectID, error) {
	for {
		ty, _, err := fetcher.Header(id)
		if err != nil {
			return oid.ObjectID{}, err
		}

		switch ty {
		case typ.Blob:
			return id, nil
		case typ.Tag:
			tag, err := fetcher.ExactTag(id)
			if err != nil {
				return oid.ObjectID{}, err
			}

			id = tag.Object().TargetID
		case typ.Unknown, typ.Commit, typ.Tree:
			return oid.ObjectID{}, &errs.ObjectTypeError{OID: id, Got: ty, Want: typ.Blob}
		default:
			return oid.ObjectID{}, &errs.ObjectTypeError{OID: id, Got: ty, Want: typ.Blob}
		}
	}
}

// PeelToBlobReader returns a reader for the content of the peeled blob at id,
// together with its content size in bytes.
//
// Labels: Life-Parent, Close-Caller.
func (fetcher *Fetcher) PeelToBlobReader(id oid.ObjectID) (io.ReadCloser, uint64, error) {
	blobID, err := fetcher.PeelToBlobID(id)
	if err != nil {
		return nil, 0, err
	}

	return fetcher.ExactBlobReader(blobID)
}
