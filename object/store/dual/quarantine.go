package dual

import (
	"context"
	"errors"
	"fmt"
	"io"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
	"lindenii.org/go/furgit/object/store/mix"
	"lindenii.org/go/furgit/object/typ"
)

// BeginObjectQuarantine begins an object-wise quarantine on the object side.
func (dual *Dual) BeginObjectQuarantine(opts store.ObjectQuarantineOptions) (store.ObjectQuarantine, error) {
	return dual.object.BeginObjectQuarantine(opts) //nolint:wrapcheck
}

// BeginPackQuarantine begins a pack-wise quarantine on the pack side.
func (dual *Dual) BeginPackQuarantine(opts store.PackQuarantineOptions) (store.PackQuarantine, error) {
	return dual.pack.BeginPackQuarantine(opts) //nolint:wrapcheck
}

// BeginCoordinatedQuarantine begins a quarantine spanning both sides.
//
// If the pack side fails to begin,
// the already-begun object side is discarded before returning.
func (dual *Dual) BeginCoordinatedQuarantine(opts store.CoordinatedQuarantineOptions) (store.CoordinatedQuarantine, error) {
	objectQ, err := dual.object.BeginObjectQuarantine(opts.Object)
	if err != nil {
		return nil, fmt.Errorf("object/store/dual: begin object quarantine: %w", err)
	}

	packQ, err := dual.pack.BeginPackQuarantine(opts.Pack)
	if err != nil {
		_ = objectQ.Discard()

		return nil, fmt.Errorf("object/store/dual: begin pack quarantine: %w", err)
	}

	return newCoordinatedQuarantine(objectQ, packQ), nil
}

// coordinatedQuarantine is one coordinated quarantine over both sides.
type coordinatedQuarantine struct {
	objectQ store.ObjectQuarantine
	packQ   store.PackQuarantine
	reader  store.ObjectReader
}

var _ store.CoordinatedQuarantine = (*coordinatedQuarantine)(nil)

func newCoordinatedQuarantine(
	objectQ store.ObjectQuarantine,
	packQ store.PackQuarantine,
) *coordinatedQuarantine {
	return &coordinatedQuarantine{
		objectQ: objectQ,
		packQ:   packQ,
		reader:  mix.New(objectQ, packQ),
	}
}

func (quarantine *coordinatedQuarantine) ReadBytesFull(id id.ObjectID) ([]byte, error) {
	return quarantine.reader.ReadBytesFull(id) //nolint:wrapcheck
}

func (quarantine *coordinatedQuarantine) ReadBytesContent(id id.ObjectID) (typ.Type, []byte, error) {
	return quarantine.reader.ReadBytesContent(id) //nolint:wrapcheck
}

func (quarantine *coordinatedQuarantine) ReadReaderFull(id id.ObjectID) (io.ReadCloser, error) {
	return quarantine.reader.ReadReaderFull(id) //nolint:wrapcheck
}

func (quarantine *coordinatedQuarantine) ReadReaderContent(id id.ObjectID) (typ.Type, int, io.ReadCloser, error) {
	return quarantine.reader.ReadReaderContent(id) //nolint:wrapcheck
}

func (quarantine *coordinatedQuarantine) ReadSize(id id.ObjectID) (int, error) {
	return quarantine.reader.ReadSize(id) //nolint:wrapcheck
}

func (quarantine *coordinatedQuarantine) ReadHeader(id id.ObjectID) (typ.Type, int, error) {
	return quarantine.reader.ReadHeader(id) //nolint:wrapcheck
}

func (quarantine *coordinatedQuarantine) Refresh() error {
	return quarantine.reader.Refresh() //nolint:wrapcheck
}

func (quarantine *coordinatedQuarantine) WriteBytesFull(raw []byte) (id.ObjectID, error) {
	return quarantine.objectQ.WriteBytesFull(raw) //nolint:wrapcheck
}

func (quarantine *coordinatedQuarantine) WriteBytesContent(ty typ.Type, content []byte) (id.ObjectID, error) {
	return quarantine.objectQ.WriteBytesContent(ty, content) //nolint:wrapcheck
}

func (quarantine *coordinatedQuarantine) WriteReaderFull(src io.Reader) (id.ObjectID, error) {
	return quarantine.objectQ.WriteReaderFull(src) //nolint:wrapcheck
}

func (quarantine *coordinatedQuarantine) WriteReaderContent(ty typ.Type, size int, src io.Reader) (id.ObjectID, error) {
	return quarantine.objectQ.WriteReaderContent(ty, size, src) //nolint:wrapcheck
}

func (quarantine *coordinatedQuarantine) WritePack(ctx context.Context, src io.Reader, opts store.PackWriteOptions) error {
	return quarantine.packQ.WritePack(ctx, src, opts) //nolint:wrapcheck
}

// Promote publishes both halves and joins their errors.
func (quarantine *coordinatedQuarantine) Promote() error {
	return errors.Join(quarantine.objectQ.Promote(), quarantine.packQ.Promote())
}

// Discard abandons both halves and joins their errors.
func (quarantine *coordinatedQuarantine) Discard() error {
	return errors.Join(quarantine.objectQ.Discard(), quarantine.packQ.Discard())
}
