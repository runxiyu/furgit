package furgit

const defaultDeltaWindow = 64

type objectToPack struct {
	id         Hash
	ty         ObjectType
	body       []byte
	offset     uint64
	deltaDepth int
	inPack     bool
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
	for i := len(ctx.candidates) - 1; i >= 0; i-- {
		base := ctx.candidates[i]
		if base.ty != ObjectTypeBlob {
			continue
		}
		if base.deltaDepth >= maxDepth {
			continue
		}
		delta, ok := deltaTry(base.body, obj.body, seed, minSavings)
		if !ok {
			continue
		}
		if bestDelta == nil || len(delta) < len(bestDelta) {
			bestDelta = delta
			bestBase = base
		}
	}
	return bestBase, bestDelta
}
