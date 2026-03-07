package service

import (
	"os"

	"codeberg.org/lindenii/furgit/objectstore"
	"codeberg.org/lindenii/furgit/objectstore/loose"
	"codeberg.org/lindenii/furgit/objectstore/memory"
	objectmix "codeberg.org/lindenii/furgit/objectstore/mix"
	"codeberg.org/lindenii/furgit/objectstore/packed"
)

func (service *Service) openQuarantinedObjects(quarantineName string) (objectstore.Store, error) {
	if quarantineName == "" {
		return memory.New(service.opts.Algorithm), nil
	}

	looseRoot, err := service.opts.ObjectsRoot.OpenRoot(quarantineName)
	if err != nil {
		return nil, err
	}

	looseStore, err := loose.New(looseRoot, service.opts.Algorithm)
	if err != nil {
		_ = looseRoot.Close()

		return nil, err
	}

	packRoot, err := looseRoot.OpenRoot("pack")
	if err == nil {
		packedStore, packedErr := packed.New(packRoot, service.opts.Algorithm)
		if packedErr != nil {
			_ = packRoot.Close()
			_ = looseStore.Close()

			return nil, packedErr
		}

		return objectmix.New(looseStore, packedStore), nil
	}

	if !os.IsNotExist(err) {
		_ = looseStore.Close()

		return nil, err
	}

	return looseStore, nil
}
