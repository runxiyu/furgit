package typ

type typeDetails struct {
	name         string
	isBaseObject bool
}

func (ty Type) details() typeDetails {
	return typeTable[ty]
}

//nolint:gochecknoglobals
var typeTable = [...]typeDetails{
	TypeInvalid:  {name: "", isBaseObject: false},
	TypeCommit:   {name: "commit", isBaseObject: true},
	TypeTree:     {name: "tree", isBaseObject: true},
	TypeBlob:     {name: "blob", isBaseObject: true},
	TypeTag:      {name: "tag", isBaseObject: true},
	TypeFuture:   {name: "", isBaseObject: false},
	TypeOfsDelta: {name: "", isBaseObject: false},
	TypeRefDelta: {name: "", isBaseObject: false},
}

//nolint:gochecknoglobals
var typeByName = map[string]Type{
	typeTable[TypeCommit].name: TypeCommit,
	typeTable[TypeTree].name:   TypeTree,
	typeTable[TypeBlob].name:   TypeBlob,
	typeTable[TypeTag].name:    TypeTag,
}
