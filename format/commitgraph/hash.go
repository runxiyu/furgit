package commitgraph

import (
	"bytes"
	"fmt"
	"io"

	"codeberg.org/lindenii/furgit/objectid"
)

// HashVersion returns the commit-graph hash version.
func (reader *Reader) HashVersion() uint8 {
	return reader.hashVersion
}

func validateChainBaseHashes(algo objectid.Algorithm, chain []string, idx int, graph *layer) error {
	if idx == 0 {
		if len(graph.chunkBaseGraphs) != 0 {
			return &ErrMalformed{Path: graph.path, Reason: "unexpected BASE chunk in first graph"}
		}

		return nil
	}

	hashSize := algo.Size()

	expectedLen := idx * hashSize
	if len(graph.chunkBaseGraphs) != expectedLen {
		return &ErrMalformed{
			Path:   graph.path,
			Reason: fmt.Sprintf("BASE chunk length %d does not match expected %d", len(graph.chunkBaseGraphs), expectedLen),
		}
	}

	for i := range idx {
		start := i * hashSize
		end := start + hashSize

		baseHash, err := objectid.FromBytes(algo, graph.chunkBaseGraphs[start:end])
		if err != nil {
			return err
		}

		if baseHash.String() != chain[i] {
			return &ErrMalformed{
				Path:   graph.path,
				Reason: fmt.Sprintf("BASE chunk mismatch at index %d", i),
			}
		}
	}

	return nil
}

func verifyTrailerHash(data []byte, algo objectid.Algorithm, path string) error {
	hashSize := algo.Size()
	if len(data) < hashSize {
		return &ErrMalformed{Path: path, Reason: "file too short for trailer"}
	}

	hashImpl, err := algo.New()
	if err != nil {
		return err
	}

	_, err = io.Copy(hashImpl, bytes.NewReader(data[:len(data)-hashSize]))
	if err != nil {
		return err
	}

	got := hashImpl.Sum(nil)

	want := data[len(data)-hashSize:]
	if !bytes.Equal(got, want) {
		return &ErrMalformed{Path: path, Reason: "trailer hash mismatch"}
	}

	return nil
}
