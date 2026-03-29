package read

// LayerInfo describes one loaded commit-graph layer.
type LayerInfo struct {
	Path      string
	BaseCount uint32
	Commits   uint32
}

// Layers returns loaded layer metadata in native chain order.
//
// Labels: MT-Safe, Life-Independent.
func (reader *Reader) Layers() []LayerInfo {
	out := make([]LayerInfo, 0, len(reader.layers))
	for i := range reader.layers {
		layer := reader.layers[i]
		out = append(out, LayerInfo{
			Path:      layer.path,
			BaseCount: layer.baseCount,
			Commits:   layer.numCommits,
		})
	}

	return out
}
