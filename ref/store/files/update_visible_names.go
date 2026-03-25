package files

func (executor *refUpdateExecutor) collectVisibleNames() (map[string]struct{}, error) {
	names := make(map[string]struct{})

	looseNames, err := executor.store.collectLooseRefNames()
	if err != nil {
		return nil, err
	}

	for _, name := range looseNames {
		names[name] = struct{}{}
	}

	packed, err := executor.store.readPackedRefs()
	if err != nil {
		return nil, err
	}

	for name := range packed.byName {
		if _, exists := names[name]; exists {
			continue
		}

		names[name] = struct{}{}
	}

	return names, nil
}
