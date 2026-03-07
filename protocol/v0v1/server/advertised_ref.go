package server

import "codeberg.org/lindenii/furgit/objectid"

// AdvertisedRef is one ref entry in one v0/v1 server advertisement.
type AdvertisedRef struct {
	// Name is the advertised reference name. It may be HEAD or one full
	// reference name.
	Name string
	// ID is the object ID currently advertised for Name.
	ID objectid.ObjectID
	// Peeled is the peeled annotated-tag target when available.
	//
	// If set, advertisement writes one immediate "<name>^{}" line after the
	// main entry, matching Git's advertisement rules.
	Peeled *objectid.ObjectID
}

// Advertisement is one server-side ref advertisement.
type Advertisement struct {
	Refs []AdvertisedRef
}
