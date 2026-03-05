package repository

import "errors"

// Close closes owned stores and filesystem roots.
// The behavior of the repo after Close is undefined.
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

	if repo.objectsLooseForWritingOnly != nil {
		err := repo.objectsLooseForWritingOnly.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
