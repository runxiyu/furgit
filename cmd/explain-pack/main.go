package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"lindenii.org/go/furgit/internal/format/packfile"
	"lindenii.org/go/furgit/internal/format/packidx"
	"lindenii.org/go/furgit/internal/mmap"
	"lindenii.org/go/furgit/internal/utils"
	"lindenii.org/go/furgit/object/id"
)

func main() {
	format := flag.String("format", "", "object format of the pack: sha1 or sha256 (required)")

	flag.Parse()

	err := run(*format, flag.Args(), os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, "explain-pack:", err)
		os.Exit(1)
	}
}

type explainer struct {
	data         []byte
	objectFormat id.ObjectFormat
	out          *bufio.Writer

	idx *packidx.Packidx

	cache    *baseCache
	oidIndex map[id.ObjectID]int
}

func run(format string, args []string, stdin io.Reader, stdout io.Writer) error {
	if format == "" {
		return fmt.Errorf("the -format flag is required (sha1 or sha256)")
	}

	objectFormat, err := id.ParseObjectFormat(format)
	if err != nil {
		return fmt.Errorf("invalid -format %q: %w", format, err)
	}

	if len(args) > 1 {
		return fmt.Errorf("at most one pack file argument is accepted, got %d", len(args))
	}

	data, idx, closers, err := openInput(args, objectFormat, stdin)
	if err != nil {
		return err
	}

	defer func() {
		for _, c := range closers {
			_ = c.Close()
		}
	}()

	out := bufio.NewWriter(stdout)

	explainer := &explainer{
		data:         data,
		objectFormat: objectFormat,
		out:          out,
		idx:          idx,
		cache:        newBaseCache(),
		oidIndex:     make(map[id.ObjectID]int),
	}

	err = explainer.explain()
	if err != nil {
		return err
	}

	return out.Flush()
}

func openInput(args []string, objectFormat id.ObjectFormat, stdin io.Reader) ([]byte, *packidx.Packidx, []io.Closer, error) {
	if len(args) == 0 {
		data, err := io.ReadAll(stdin)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("reading pack from stdin: %w", err)
		}

		return data, nil, nil, nil
	}

	packPath := args[0]

	packMapping, err := mapPath(packPath)
	if err != nil {
		return nil, nil, nil, err
	}

	closers := []io.Closer{packMapping}

	idx, idxMapping, err := openIndex(packPath, objectFormat)
	if err != nil {
		_ = packMapping.Close()

		return nil, nil, nil, err
	}

	if idxMapping != nil {
		closers = append(closers, idxMapping)
	}

	return packMapping.Data(), idx, closers, nil
}

func openIndex(packPath string, objectFormat id.ObjectFormat) (*packidx.Packidx, *mmap.Mmap, error) {
	idxPath := strings.TrimSuffix(packPath, ".pack") + ".idx"

	file, err := os.Open(idxPath) //#nosec G304
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}

		return nil, nil, fmt.Errorf("opening index %q: %w", idxPath, err)
	}

	defer func() { _ = file.Close() }()

	mapping, err := mmap.Open(file)
	if err != nil {
		return nil, nil, fmt.Errorf("mapping index %q: %w", idxPath, err)
	}

	idx, err := packidx.Parse(mapping.Data(), objectFormat.Size())
	if err != nil {
		_ = mapping.Close()

		return nil, nil, fmt.Errorf("parsing index %q: %w", idxPath, err)
	}

	return &idx, mapping, nil
}

func mapPath(path string) (*mmap.Mmap, error) {
	file, err := os.Open(path) //#nosec G304
	if err != nil {
		return nil, fmt.Errorf("opening pack %q: %w", path, err)
	}

	defer func() { _ = file.Close() }()

	mapping, err := mmap.Open(file)
	if err != nil {
		return nil, fmt.Errorf("mapping pack %q: %w", path, err)
	}

	return mapping, nil
}

func (explainer *explainer) printf(format string, args ...any) {
	utils.BestEffortFprintf(explainer.out, format, args...)
}

func (explainer *explainer) explain() error {
	hashSize := explainer.objectFormat.Size()

	if len(explainer.data) < packfile.HeaderLen+hashSize {
		return fmt.Errorf("pack is too short to contain a header and a %d-byte trailer", hashSize)
	}

	count, err := explainer.explainHeader()
	if err != nil {
		return err
	}

	cursor := packfile.HeaderLen

	for num := 1; num <= count; num++ {
		next, err := explainer.explainEntry(num, count, cursor)
		if err != nil {
			return err
		}

		cursor = next
	}

	return explainer.explainTrailer(cursor)
}

func (explainer *explainer) explainHeader() (int, error) {
	header, err := packfile.ParseHeader(explainer.data[:packfile.HeaderLen])
	if err != nil {
		return 0, fmt.Errorf("pack header: %w", err)
	}

	explainer.printf("pack header\n")
	explainer.printf("\tmagic\t\"PACK\"\n")
	explainer.printf("\tversion\t2\n")
	explainer.printf("\tobjects\t%d\n", header.ObjectCount)
	explainer.printf("\n")

	return int(header.ObjectCount), nil
}

func (explainer *explainer) explainTrailer(cursor int) error {
	hashSize := explainer.objectFormat.Size()
	trailerStart := len(explainer.data) - hashSize

	if cursor != trailerStart {
		explainer.printf(
			"note\t%d byte(s) between the last entry and the trailer were unaccounted for\n",
			trailerStart-cursor,
		)
	}

	trailer := explainer.data[trailerStart:]

	explainer.printf("pack trailer\n")
	explainer.printf("\tchecksum\t%s\n", hex.EncodeToString(trailer))

	hashImpl, err := explainer.objectFormat.New()
	if err != nil {
		return fmt.Errorf("object/store: %w", err)
	}

	_, _ = hashImpl.Write(explainer.data[:trailerStart])

	if bytes.Equal(hashImpl.Sum(nil), trailer) {
		explainer.printf("\trecomputed\tmatches\n")
	} else {
		explainer.printf("\trecomputed\tMISMATCH (corrupt pack or wrong -format)\n")
	}

	return nil
}
