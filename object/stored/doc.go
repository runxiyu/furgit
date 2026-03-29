// Package stored wraps parsed objects with the object IDs they were loaded
// under.
//
// Parsed Git object values do not carry storage identity on their own. This
// package provides a small generic wrapper for the common case where callers
// need both the parsed object value and the object ID it was read from.
package stored
