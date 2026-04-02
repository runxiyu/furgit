package intconv

import "errors"

// ErrOverflow indicates that
// a requested checked integer conversion
// is impossible due to
// integer overflow,
// integer underflow,
// unexpected negative integers,
// or a similar situation.
var ErrOverflow = errors.New("intconv: overflow")
