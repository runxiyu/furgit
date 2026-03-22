package read

// Close releases all mapped commit-graph files.
//
// Repeated calls to Close are undefined behavior.
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
