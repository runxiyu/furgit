package packed_test

import (
	"errors"
	"testing"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
)

func TestLookupMissing(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			_, prefix, seeded := makeGitPack(t, objectFormat)
			packedStore := openPackedStore(t, prefix, objectFormat)

			raw := seeded.Blobs[0].Bytes()
			raw[len(raw)-1] ^= 0xff

			missing, err := objectFormat.FromBytes(raw)
			if err != nil {
				t.Fatalf("FromBytes: %v", err)
			}

			_, _, err = packedStore.ReadBytesContent(missing)
			if !errors.Is(err, store.ErrObjectNotFound) {
				t.Fatalf("ReadBytesContent error = %v, want ErrObjectNotFound", err)
			}
		})
	}
}
