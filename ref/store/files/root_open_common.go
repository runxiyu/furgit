package files

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func openCommonRoot(gitRoot *os.Root) (*os.Root, error) {
	content, err := gitRoot.ReadFile("commondir")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return gitRoot.OpenRoot(".")
		}

		return nil, err
	}

	commonDir := strings.TrimSpace(string(content))
	if commonDir == "" {
		return nil, os.ErrNotExist
	}

	if filepath.IsAbs(commonDir) {
		return os.OpenRoot(commonDir)
	}

	// This is okay because that's how Git defines it anyway.
	return os.OpenRoot(filepath.Join(gitRoot.Name(), commonDir))
}
