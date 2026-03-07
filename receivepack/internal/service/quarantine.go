package service

import (
	"crypto/rand"
	"os"
)

// createQuarantineRoot creates one per-push quarantine directory beneath the
// permanent objects root.
func (service *Service) createQuarantineRoot() (string, *os.Root, error) {
	name := "tmp_objdir-incoming-" + rand.Text()

	err := service.opts.ObjectsRoot.Mkdir(name, 0o700)
	if err != nil {
		return "", nil, err
	}

	root, err := service.opts.ObjectsRoot.OpenRoot(name)
	if err != nil {
		_ = service.opts.ObjectsRoot.RemoveAll(name)

		return "", nil, err
	}

	return name, root, nil
}
