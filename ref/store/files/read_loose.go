package files

import (
	"errors"
	"fmt"
	"os"
	"strings"

	objectid "lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/ref"
	refstore "lindenii.org/go/furgit/ref/store"
)

func (store *Store) readLooseRef(name string) (ref.Ref, error) { //nolint:ireturn
	refPath := store.loosePath(name)

	data, err := store.rootFor(refPath.root).ReadFile(refPath.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, refstore.ErrReferenceNotFound
		}

		return nil, err
	}

	line := strings.TrimRightFunc(string(data), isRefWhitespace)
	if strings.HasPrefix(line, "ref:") {
		target := strings.TrimLeftFunc(line[len("ref:"):], isRefWhitespace)
		if target == "" {
			return nil, brokenRefError{name: name, err: fmt.Errorf("empty symbolic target")}
		}

		return ref.Symbolic{
			RefName: name,
			Target:  target,
		}, nil
	}

	id, err := objectid.ParseHex(store.algo, line)
	if err != nil {
		return nil, brokenRefError{name: name, err: err}
	}

	return ref.Detached{
		RefName: name,
		ID:      id,
	}, nil
}
