package receivepack

import (
	"errors"

	common "lindenii.org/go/furgit/network/protocol/v0v1/server"
	"lindenii.org/go/furgit/ref"
	refstore "lindenii.org/go/furgit/ref/store"
)

func advertisedRefs(opts Options) ([]common.AdvertisedRef, error) {
	listed, err := opts.Refs.List("")
	if err != nil {
		return nil, err
	}

	return buildAdvertisedRefs(opts, listed)
}

func buildAdvertisedRefs(opts Options, listed []ref.Ref) ([]common.AdvertisedRef, error) {
	refs := make([]common.AdvertisedRef, 0, len(listed))
	for _, entry := range listed {
		switch resolved := entry.(type) {
		case ref.Detached:
			advertised := common.AdvertisedRef{
				Name: resolved.Name(),
				ID:   resolved.ID,
			}

			if resolved.Peeled != nil {
				advertised.Peeled = resolved.Peeled
			}

			refs = append(refs, advertised)
		case ref.Symbolic:
			if resolved.Name() != "HEAD" {
				continue
			}

			head, err := opts.Refs.ResolveToDetached("HEAD")
			if err != nil {
				if errors.Is(err, refstore.ErrReferenceNotFound) {
					continue
				}

				return nil, err
			}

			refs = append(refs, common.AdvertisedRef{
				Name: "HEAD",
				ID:   head.ID,
			})
		}
	}

	return refs, nil
}
