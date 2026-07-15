package files

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/ref"
	refname "lindenii.org/go/furgit/ref/name"
	"lindenii.org/go/furgit/ref/store"
)

const looseReadAttempts = 3

// errBrokenRef indicates that an on-disk reference exists
// but its content cannot be parsed.
var errBrokenRef = errors.New("ref/store/files: broken reference")

// errRefDirectory indicates that a loose reference path is a directory,
// possibly shadowing a packed reference.
var errRefDirectory = errors.New("ref/store/files: reference path is a directory")

// errUnstableRef indicates that a loose reference kept changing shape
// between snapshots while being read.
var errUnstableRef = errors.New("ref/store/files: reference changed repeatedly during read")

// Resolve resolves one reference name from the visible namespace,
// reading the loose reference first
// and falling back to packed-refs when the loose reference is missing
// or shadowed by a directory.
func (files *Files) Resolve(name string) (ref.Ref, error) {
	if name == "" {
		return nil, store.ErrReferenceNotFound
	}

	resolved, err := files.readLooseRef(name)
	if err == nil {
		return resolved, nil
	}

	if !errors.Is(err, store.ErrReferenceNotFound) && !errors.Is(err, errRefDirectory) {
		return nil, err
	}

	packed, packedErr := files.readPackedRefs()
	if packedErr != nil {
		return nil, packedErr
	}

	direct, ok := packed.byName[name]
	if !ok {
		return nil, store.ErrReferenceNotFound
	}

	return direct, nil
}

// ResolveToDirect resolves symbolic references
// until one direct reference is reached.
func (files *Files) ResolveToDirect(name string) (ref.Direct, error) {
	cur := name
	seen := make(map[string]struct{})

	for {
		if _, ok := seen[cur]; ok {
			return ref.Direct{}, fmt.Errorf("%w: at %q", store.ErrSymbolicCycle, cur)
		}

		seen[cur] = struct{}{}

		resolved, err := files.Resolve(cur)
		if err != nil {
			return ref.Direct{}, err
		}

		switch resolved := resolved.(type) {
		case ref.Direct:
			return resolved, nil
		case ref.Symbolic:
			if resolved.Target == "" {
				return ref.Direct{}, fmt.Errorf(
					"%w: symbolic reference %q has empty target",
					store.ErrInvalidValue, resolved.Name(),
				)
			}

			cur = resolved.Target
		default:
			panic(fmt.Sprintf("ref/store/files: unsupported reference type %T", resolved))
		}
	}
}

// readLooseRef reads one loose reference file.
//
// A symbolic-link reference whose target is a valid name under refs/
// is reported as a symbolic reference;
// other symbolic links are read through.
func (files *Files) readLooseRef(name string) (ref.Ref, error) {
	loc := files.loosePath(name)
	root := files.root(loc.kind)

	for range looseReadAttempts {
		info, err := root.Lstat(loc.path)
		if err != nil {
			// A non-directory in the path means the reference
			// cannot exist loose,
			// such as refs/heads/a shadowing refs/heads/a/b.
			if errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ENOTDIR) {
				return nil, store.ErrReferenceNotFound
			}

			return nil, fmt.Errorf("ref/store/files: lstat %q: %w", loc.path, err)
		}

		if info.Mode()&os.ModeSymlink != 0 {
			target, linkErr := root.Readlink(loc.path)
			if linkErr != nil {
				if errors.Is(linkErr, os.ErrNotExist) || errors.Is(linkErr, syscall.EINVAL) {
					continue
				}

				return nil, fmt.Errorf("ref/store/files: readlink %q: %w", loc.path, linkErr)
			}

			if strings.HasPrefix(target, "refs/") &&
				refname.Validate(target, refname.Options{AllowOneLevel: false, RefspecPattern: false}) == nil {
				return ref.Symbolic{RefName: name, Target: target}, nil
			}
		}

		if info.IsDir() {
			return nil, errRefDirectory
		}

		data, err := root.ReadFile(loc.path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}

			return nil, fmt.Errorf("ref/store/files: read %q: %w", loc.path, err)
		}

		return parseLooseRef(files.objectFormat, name, data)
	}

	return nil, errUnstableRef
}

// parseLooseRef parses loose reference file content.
//
// A "ref:" prefix introduces a symbolic reference
// whose target is the remaining content
// after leading and trailing whitespace.
// Anything else must begin with one object ID
// followed only by whitespace or the end of the content;
// data after the whitespace is ignored,
// which lets pseudorefs with trailing metadata read as direct references.
func parseLooseRef(objectFormat id.ObjectFormat, name string, data []byte) (ref.Ref, error) {
	content := string(data)

	if rest, ok := strings.CutPrefix(content, "ref:"); ok {
		target := strings.TrimFunc(rest, isRefWhitespace)
		if target == "" {
			return nil, fmt.Errorf("%w: %q: empty symbolic target", errBrokenRef, name)
		}

		return ref.Symbolic{RefName: name, Target: target}, nil
	}

	hexLen := objectFormat.HexLen()
	if len(content) < hexLen {
		return nil, fmt.Errorf("%w: %q: content too short", errBrokenRef, name)
	}

	oid, err := objectFormat.FromString(content[:hexLen])
	if err != nil {
		return nil, fmt.Errorf("%w: %q: %w", errBrokenRef, name, err)
	}

	if len(content) > hexLen && !isRefWhitespace(rune(content[hexLen])) {
		return nil, fmt.Errorf("%w: %q: trailing data after object id", errBrokenRef, name)
	}

	return ref.Direct{
		RefName:   name,
		ID:        oid,
		PeelState: ref.PeelUnknown,
		PeeledID:  id.ObjectID{},
	}, nil
}

func isRefWhitespace(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\r', '\v', '\f':
		return true
	default:
		return false
	}
}
