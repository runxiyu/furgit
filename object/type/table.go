package objecttype

//nolint:gochecknoglobals
var typeTable = [...]typeDetails{
	TypeInvalid:  {},
	TypeCommit:   {name: "commit", isBaseObject: true},
	TypeTree:     {name: "tree", isBaseObject: true},
	TypeBlob:     {name: "blob", isBaseObject: true},
	TypeTag:      {name: "tag", isBaseObject: true},
	TypeFuture:   {},
	TypeOfsDelta: {},
	TypeRefDelta: {},
}

//nolint:gochecknoglobals
var typeByName = map[string]Type{
	typeTable[TypeCommit].name: TypeCommit,
	typeTable[TypeTree].name:   TypeTree,
	typeTable[TypeBlob].name:   TypeBlob,
	typeTable[TypeTag].name:    TypeTag,
}
