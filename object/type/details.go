package objecttype

type typeDetails struct {
	name         string
	isBaseObject bool
}

func (ty Type) details() typeDetails {
	return typeTable[ty]
}
