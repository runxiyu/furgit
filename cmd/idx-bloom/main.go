package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"lindenii.org/go/furgit/internal/format/packidx"
	"lindenii.org/go/furgit/internal/format/packidx/bloom"
	"lindenii.org/go/furgit/object/id"
)

func main() {
	format := flag.String("format", "", "object format of the index: sha1 or sha256 (required)")

	flag.Parse()

	err := run(*format, flag.Args(), os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, "idx-bloom:", err)
		os.Exit(1)
	}
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
		return fmt.Errorf("at most one index file argument is accepted, got %d", len(args))
	}

	data, err := readInput(args, stdin)
	if err != nil {
		return err
	}

	index, err := packidx.Parse(data, objectFormat.Size())
	if err != nil {
		return fmt.Errorf("parsing index: %w", err)
	}

	filter, err := buildFilter(objectFormat, &index)
	if err != nil {
		return err
	}

	_, err = stdout.Write(filter)
	if err != nil {
		return fmt.Errorf("writing filter: %w", err)
	}

	return nil
}

func readInput(args []string, stdin io.Reader) ([]byte, error) {
	if len(args) == 0 {
		data, err := io.ReadAll(stdin)
		if err != nil {
			return nil, fmt.Errorf("reading index from stdin: %w", err)
		}

		return data, nil
	}

	data, err := os.ReadFile(args[0]) //#nosec G304
	if err != nil {
		return nil, fmt.Errorf("reading index %q: %w", args[0], err)
	}

	return data, nil
}

func buildFilter(objectFormat id.ObjectFormat, index *packidx.Packidx) ([]byte, error) {
	objects := index.NumObjects()

	bucketCount, k, err := bloom.RecommendParams(objectFormat, objects)
	if err != nil {
		return nil, fmt.Errorf("choosing parameters: %w", err)
	}

	builder, err := bloom.NewBuilder(objectFormat, bucketCount, k, index.PackHash())
	if err != nil {
		return nil, fmt.Errorf("creating builder: %w", err)
	}

	for pos := range objects {
		builder.Add(index.OIDAt(pos))
	}

	return builder.Bytes(), nil
}
