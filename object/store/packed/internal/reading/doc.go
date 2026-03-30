// Package reading implements the packed-store read path: pack and index
// discovery, lookup, caching, and object reads from existing packfiles.
//
// Obviously, this internal package is not meant to be used by anyone
// other than object/store/packed.
package reading
