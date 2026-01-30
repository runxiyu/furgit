package furgit

const defaultDeltaWindow = 64

type objectToPack struct {
	id         Hash
	ty         ObjectType
	body       []byte
	size       int
	offset     uint64
	deltaDepth int
	inPack     bool
	preferred  bool
}

type deltaContext struct {
	window     int
	candidates []*objectToPack
}

func (ctx *deltaContext) addCandidate(obj *objectToPack) {
	if ctx.window <= 0 {
		return
	}
	ctx.candidates = append(ctx.candidates, obj)
	if len(ctx.candidates) > ctx.window {
		over := len(ctx.candidates) - ctx.window
		ctx.candidates = ctx.candidates[over:]
	}
}

func pickDeltaBase(ctx *deltaContext, obj *objectToPack, seed uint64, minSavings, maxDepth int) (*objectToPack, []byte) {
	if ctx == nil || len(ctx.candidates) == 0 {
		return nil, nil
	}
	if maxDepth <= 0 {
		maxDepth = 1
	}
	var bestBase *objectToPack
	var bestDelta []byte
	var bestPreferred *objectToPack
	var bestPreferredDelta []byte
	for i := len(ctx.candidates) - 1; i >= 0; i-- {
		base := ctx.candidates[i]
		if base.ty != ObjectTypeBlob {
			continue
		}
		if base.deltaDepth >= maxDepth {
			continue
		}
		if !deltaSizeOk(base, obj, maxDepth) {
			continue
		}
		delta, ok := deltaTry(base.body, obj.body, seed, minSavings)
		if base.preferred {
			delta, ok = deltaTry(base.body, obj.body, seed, 0)
		}
		if !ok {
			continue
		}
		if base.preferred {
			if bestPreferredDelta == nil || len(delta) < len(bestPreferredDelta) {
				bestPreferredDelta = delta
				bestPreferred = base
			}
			continue
		}
		if bestDelta == nil || len(delta) < len(bestDelta) {
			bestDelta = delta
			bestBase = base
		}
	}
	if bestPreferred != nil {
		return bestPreferred, bestPreferredDelta
	}
	return bestBase, bestDelta
}

func deltaSizeOk(base, target *objectToPack, maxDepth int) bool {
	if base == nil || target == nil {
		return false
	}
	if base.size <= 0 || target.size <= 0 {
		return false
	}
	if maxDepth <= 0 {
		maxDepth = 1
	}
	if base.deltaDepth >= maxDepth {
		return false
	}
	if target.size < base.size/32 {
		return false
	}
	maxSize := target.size/2 - 32
	if maxSize <= 0 {
		return false
	}
	sizediff := 0
	if base.size < target.size {
		sizediff = target.size - base.size
	}
	if sizediff >= maxSize {
		return false
	}
	return true
}
