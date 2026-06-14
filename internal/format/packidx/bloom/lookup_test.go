package bloom_test

import (
	"testing"

	"lindenii.org/go/furgit/internal/format/packidx/bloom"
	"lindenii.org/go/furgit/object/id"
)

func TestMayContainBadLength(t *testing.T) {
	t.Parallel()

	format := id.ObjectFormatSHA256

	builder, err := bloom.NewBuilder(format, 4, 2)
	if err != nil {
		t.Fatal(err)
	}

	filter, err := bloom.Parse(builder.Bytes(), format)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if recover() == nil {
			t.Fatal("MayContain did not panic on a short object ID")
		}
	}()

	filter.MayContain(make([]byte, format.Size()-1))
}
