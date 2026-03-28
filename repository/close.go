package repository

import "errors"

// Close closes repository-owned stores and filesystem roots.
//
// Labels: MT-Unsafe.
func (repo *Repository) Close() error {
	var errs []error

	if repo.refs != nil {
		err := repo.refs.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	if repo.objects != nil {
		err := repo.objects.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	if repo.objectsPacked != nil {
		err := repo.objectsPacked.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	if repo.objectsLoose != nil {
		err := repo.objectsLoose.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	if repo.objectsPackRoot != nil {
		err := repo.objectsPackRoot.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	if repo.objectsRoot != nil {
		err := repo.objectsRoot.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	if repo.refRoot != nil {
		err := repo.refRoot.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
