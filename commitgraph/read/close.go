package read

// Close releases all mapped commit-graph files.
func (reader *Reader) Close() error {
	var closeErr error

	for i := len(reader.layers) - 1; i >= 0; i-- {
		err := reader.layers[i].close()
		if err != nil && closeErr == nil {
			closeErr = err
		}
	}

	reader.layers = nil
	reader.total = 0

	return closeErr
}
