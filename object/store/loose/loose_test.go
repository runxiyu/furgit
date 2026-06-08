package loose_test

import (
	"errors"
	"os"
	"testing"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store/loose"
)

func TestNewRejectsUnknownObjectFormat(t *testing.T) {
	t.Parallel()

	root, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatalf("OpenRoot: %v", err)
	}

	defer func() { _ = root.Close() }()

	_, err = loose.New(root, id.ObjectFormatUnknown)
	if !errors.Is(err, id.ErrInvalidObjectFormat) {
		t.Fatalf("loose.New(unknown) = %v, want ErrInvalidObjectFormat", err)
	}
}
