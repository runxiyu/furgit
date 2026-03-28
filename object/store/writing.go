package objectstore

// There is currently no writing-store interface because different
// object store backends have very different models for writing.
// For example, a loose object store can trivially write single loose
// objects, but writing individual objects to a packfile store would
// be extremely wasteful.
//
// At some time, we will have writing-store interfaces.
