package typ

type typeDetails struct {
	name string
}

func (ty Type) details() typeDetails {
	return typeTable[ty]
}

//nolint:gochecknoglobals
var typeTable = [...]typeDetails{
	TypeCommit:  {name: "commit"},
	TypeTree:    {name: "tree"},
	TypeBlob:    {name: "blob"},
	TypeTag:     {name: "tag"},
}

//nolint:gochecknoglobals
var typeByName = map[string]Type{
	typeTable[TypeCommit].name: TypeCommit,
	typeTable[TypeTree].name:   TypeTree,
	typeTable[TypeBlob].name:   TypeBlob,
	typeTable[TypeTag].name:    TypeTag,
}
