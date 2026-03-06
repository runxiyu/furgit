package commitgraph

import "syscall"

func closeLayers(layers []layer) {
	for i := len(layers) - 1; i >= 0; i-- {
		_ = layers[i].close()
	}
}

func (layer *layer) close() error {
	var closeErr error

	if layer.data != nil {
		err := syscall.Munmap(layer.data)
		if err != nil {
			closeErr = err
		}

		layer.data = nil
	}

	if layer.file != nil {
		err := layer.file.Close()
		if err != nil && closeErr == nil {
			closeErr = err
		}

		layer.file = nil
	}

	return closeErr
}
