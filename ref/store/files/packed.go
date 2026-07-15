package files

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/ref"
)

const packedRefsHeaderPrefix = "# pack-refs with: "

// errInvalidPackedRefs indicates that the packed-refs file is malformed.
var errInvalidPackedRefs = errors.New("ref/store/files: invalid packed-refs")

// packedRefs is one parsed packed-refs file.
type packedRefs struct {
	byName  map[string]ref.Direct
	ordered []ref.Direct
	traits  packedTraits
}

// packedTraits records the peel promises declared by a packed-refs header.
//
// They are propagated verbatim on rewrite,
// because the store cannot peel objects itself yet.
// This is a TODO.
type packedTraits struct {
	// peeled means references under refs/tags/ that can be peeled
	// carry an explicit peel line;
	// absence of one there means the reference is not peelable.
	peeled bool

	// fullyPeeled extends the peeled promise to all references in the file.
	fullyPeeled bool
}

// readPackedRefs reads and parses the packed-refs file.
//
// A missing file yields an empty result.
func (files *Files) readPackedRefs() (*packedRefs, error) {
	file, err := files.commonRoot.Open("packed-refs")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &packedRefs{
				byName:  make(map[string]ref.Direct),
				ordered: nil,
				traits:  packedTraits{peeled: false, fullyPeeled: false},
			}, nil
		}

		return nil, fmt.Errorf("ref/store/files: open packed-refs: %w", err)
	}

	defer func() { _ = file.Close() }()

	return parsePackedRefs(file, files.objectFormat)
}

// parsePackedRefs parses packed-refs content.
func parsePackedRefs(reader io.Reader, objectFormat id.ObjectFormat) (*packedRefs, error) {
	parser := packedRefsParser{
		packed: &packedRefs{
			byName:  make(map[string]ref.Direct),
			ordered: make([]ref.Direct, 0, 32),
			traits:  packedTraits{peeled: false, fullyPeeled: false},
		},
		objectFormat: objectFormat,
		prev:         -1,
		lineNum:      0,
	}

	buffered := bufio.NewReader(reader)

	for {
		line, err := buffered.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("ref/store/files: read packed-refs: %w", err)
		}

		if line == "" && errors.Is(err, io.EOF) {
			break
		}

		parser.lineNum++

		if !strings.HasSuffix(line, "\n") {
			return nil, fmt.Errorf("%w: line %d: unterminated line", errInvalidPackedRefs, parser.lineNum)
		}

		lineErr := parser.parseLine(strings.TrimSuffix(line, "\n"))
		if lineErr != nil {
			return nil, lineErr
		}

		if errors.Is(err, io.EOF) {
			break
		}
	}

	parser.applyTraitDefaults()

	return parser.packed, nil
}

type packedRefsParser struct {
	packed       *packedRefs
	objectFormat id.ObjectFormat

	// prev indexes the most recent entry in packed.ordered,
	// which a peel line applies to.
	prev int

	lineNum int
}

func (parser *packedRefsParser) parseLine(line string) error {
	if strings.HasPrefix(line, "#") {
		return parser.parseHeader(line)
	}

	if strings.HasPrefix(line, "^") {
		return parser.parsePeel(line)
	}

	return parser.parseEntry(line)
}

func (parser *packedRefsParser) parseHeader(line string) error {
	if parser.lineNum != 1 {
		return fmt.Errorf("%w: line %d: unexpected comment line", errInvalidPackedRefs, parser.lineNum)
	}

	traitText, ok := strings.CutPrefix(line, packedRefsHeaderPrefix)
	if !ok {
		return fmt.Errorf("%w: line %d: malformed header", errInvalidPackedRefs, parser.lineNum)
	}

	for trait := range strings.FieldsSeq(traitText) {
		switch trait {
		case "peeled":
			parser.packed.traits.peeled = true
		case "fully-peeled":
			parser.packed.traits.fullyPeeled = true
		case "sorted":
		default:
		}
	}

	return nil
}

func (parser *packedRefsParser) parsePeel(line string) error {
	if parser.prev < 0 {
		return fmt.Errorf("%w: line %d: peel line without preceding entry", errInvalidPackedRefs, parser.lineNum)
	}

	if len(line) != parser.objectFormat.HexLen()+1 {
		return fmt.Errorf("%w: line %d: malformed peel line", errInvalidPackedRefs, parser.lineNum)
	}

	peeled, err := parser.objectFormat.FromString(line[1:])
	if err != nil {
		return fmt.Errorf("%w: line %d: invalid peeled object id: %w", errInvalidPackedRefs, parser.lineNum, err)
	}

	cur := parser.packed.ordered[parser.prev]
	if cur.PeelState == ref.PeelTo {
		return fmt.Errorf("%w: line %d: duplicate peel line", errInvalidPackedRefs, parser.lineNum)
	}

	cur.PeelState = ref.PeelTo
	cur.PeeledID = peeled
	parser.packed.ordered[parser.prev] = cur
	parser.packed.byName[cur.RefName] = cur

	return nil
}

func (parser *packedRefsParser) parseEntry(line string) error {
	hexLen := parser.objectFormat.HexLen()
	if len(line) < hexLen+2 || line[hexLen] != ' ' {
		return fmt.Errorf("%w: line %d: malformed entry", errInvalidPackedRefs, parser.lineNum)
	}

	oid, err := parser.objectFormat.FromString(line[:hexLen])
	if err != nil {
		return fmt.Errorf("%w: line %d: invalid object id: %w", errInvalidPackedRefs, parser.lineNum, err)
	}

	name := line[hexLen+1:]
	if _, exists := parser.packed.byName[name]; exists {
		return fmt.Errorf("%w: line %d: duplicate reference %q", errInvalidPackedRefs, parser.lineNum, name)
	}

	direct := ref.Direct{
		RefName:   name,
		ID:        oid,
		PeelState: ref.PeelUnknown,
		PeeledID:  id.ObjectID{},
	}

	parser.packed.ordered = append(parser.packed.ordered, direct)
	parser.prev = len(parser.packed.ordered) - 1
	parser.packed.byName[name] = direct

	return nil
}

// applyTraitDefaults resolves the peel state of entries without peel lines
// according to the header traits.
func (parser *packedRefsParser) applyTraitDefaults() {
	for i, entry := range parser.packed.ordered {
		if entry.PeelState == ref.PeelTo {
			continue
		}

		if parser.packed.traits.fullyPeeled ||
			(parser.packed.traits.peeled && strings.HasPrefix(entry.RefName, "refs/tags/")) {
			entry.PeelState = ref.PeelNone
			parser.packed.ordered[i] = entry
			parser.packed.byName[entry.RefName] = entry
		}
	}
}
