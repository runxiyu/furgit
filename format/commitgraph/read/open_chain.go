package read

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"codeberg.org/lindenii/furgit/internal/intconv"
	"codeberg.org/lindenii/furgit/objectid"
)

func openChain(root *os.Root, algo objectid.Algorithm) (*Reader, error) {
	chainPath := "info/commit-graphs/commit-graph-chain"

	file, err := root.Open(chainPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, &ErrMalformed{Path: chainPath, Reason: "missing commit-graph-chain"}
		}

		return nil, err
	}

	scanner := bufio.NewScanner(file)
	hashes := make([]string, 0)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		hashes = append(hashes, line)
	}

	scanErr := scanner.Err()
	closeErr := file.Close()

	if scanErr != nil {
		return nil, scanErr
	}

	if closeErr != nil {
		return nil, closeErr
	}

	if len(hashes) == 0 {
		return nil, &ErrMalformed{Path: chainPath, Reason: "empty chain"}
	}

	layers := make([]layer, 0, len(hashes))

	var total uint32

	hashVersion, err := intconv.Uint32ToUint8(algo.PackHashID())
	if err != nil {
		return nil, err
	}

	for i, hashHex := range hashes {
		expectedBaseCount, err := intconv.IntToUint32(i)
		if err != nil {
			closeLayers(layers)

			return nil, err
		}

		if len(hashHex) != algo.HexLen() {
			closeLayers(layers)

			return nil, &ErrMalformed{
				Path:   chainPath,
				Reason: fmt.Sprintf("invalid graph hash length at line %d", i+1),
			}
		}

		relPath := fmt.Sprintf("info/commit-graphs/graph-%s.graph", hashHex)

		loaded, loadErr := openLayer(root, relPath, algo)
		if loadErr != nil {
			closeLayers(layers)

			return nil, loadErr
		}

		if loaded.baseCount != expectedBaseCount {
			_ = loaded.close()

			closeLayers(layers)

			return nil, &ErrMalformed{
				Path:   relPath,
				Reason: fmt.Sprintf("BASE count %d does not match chain depth %d", loaded.baseCount, i),
			}
		}

		validateErr := validateChainBaseHashes(algo, hashes, i, loaded)
		if validateErr != nil {
			_ = loaded.close()

			closeLayers(layers)

			return nil, validateErr
		}

		loaded.globalFrom = total
		loaded.baseCount = expectedBaseCount

		totalNext := total + loaded.numCommits
		if totalNext < total {
			_ = loaded.close()

			closeLayers(layers)

			return nil, &ErrMalformed{Path: relPath, Reason: "total commit count overflow"}
		}

		total = totalNext

		layers = append(layers, *loaded)
	}

	out := &Reader{
		algo:        algo,
		hashVersion: hashVersion,
		layers:      layers,
		total:       total,
	}

	return out, nil
}
