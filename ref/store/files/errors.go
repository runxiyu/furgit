package files

import "errors"

// errBrokenRef indicates that an on-disk reference exists
// but its content cannot be parsed.
var errBrokenRef = errors.New("ref/store/files: broken reference")

// errInvalidPackedRefs indicates that the packed-refs file is malformed.
var errInvalidPackedRefs = errors.New("ref/store/files: invalid packed-refs")

// errRefDirectory indicates that a loose reference path is a directory,
// possibly shadowing a packed reference.
var errRefDirectory = errors.New("ref/store/files: reference path is a directory")

// errUnstableRef indicates that a loose reference kept changing shape
// between snapshots while being read.
var errUnstableRef = errors.New("ref/store/files: reference changed repeatedly during read")
