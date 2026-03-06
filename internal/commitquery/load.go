package commitquery

// ensureLoaded completes one node's metadata load if it has not been loaded yet.
func (ctx *Context) ensureLoaded(idx NodeIndex) error {
	if ctx.nodes[idx].loaded {
		return nil
	}

	if ctx.nodes[idx].hasGraphPos {
		return ctx.loadByGraphPos(idx)
	}

	return ctx.loadByOID(idx)
}
