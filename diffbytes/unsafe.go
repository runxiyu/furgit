package diffbytes

import "unsafe"

// // stringToBytes converts a string to a byte slice without copying the string.
// // Memory is borrowed from the string.
// // The resulting byte slice must not be modified in any form.
// func stringToBytes(s string) (bytes []byte) {
// 	return unsafe.Slice(unsafe.StringData(s), len(s)) //#nosec G103
// }

// bytesToString converts a byte slice to a string without copying the bytes.
// Memory is borrowed from the byte slice.
// The source byte slice must not be modified.
func bytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b)) //#nosec G103
}
