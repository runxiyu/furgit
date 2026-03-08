package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
)

func startCPUProfile(path string) (func() error, error) {
	//#nosec G304
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("create %q: %w", path, err)
	}

	err = pprof.StartCPUProfile(file)
	if err != nil {
		_ = file.Close()

		return nil, fmt.Errorf("start cpu profile %q: %w", path, err)
	}

	return func() error {
		pprof.StopCPUProfile()

		err := file.Close()
		if err != nil {
			return fmt.Errorf("close cpu profile %q: %w", path, err)
		}

		return nil
	}, nil
}

func writeMemProfile(path string) error {
	//#nosec G304
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %q: %w", path, err)
	}

	runtime.GC()

	err = pprof.WriteHeapProfile(file)
	if err != nil {
		_ = file.Close()

		return fmt.Errorf("write heap profile %q: %w", path, err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("close heap profile %q: %w", path, err)
	}

	return nil
}
