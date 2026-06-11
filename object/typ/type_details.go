package typ

type typeDetails struct {
	name string
}

func (ty Type) details() typeDetails {
	return typeTable[ty]
}

//nolint:gochecknoglobals
var typeTable = [...]typeDetails{
	Commit: {name: "commit"},
	Tree:   {name: "tree"},
	Blob:   {name: "blob"},
	Tag:    {name: "tag"},
}

//nolint:gochecknoglobals
var typeByName = map[string]Type{
	typeTable[Commit].name: Commit,
	typeTable[Tree].name:   Tree,
	typeTable[Blob].name:   Blob,
	typeTable[Tag].name:    Tag,
}
