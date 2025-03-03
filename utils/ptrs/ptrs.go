// Package ptrs provides utilities for working with pointers.
package ptrs

var (
	TrueFloat  = Ptr(1.0)
	FalseFloat = Ptr(0.0)
)

// String returns a pointer to a value.
func Ptr[T any](v T) *T {
	return &v
}

// StringNil returns a pointer to a string.
// If s is empty, it returns nil.
func StringNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Int64Nil returns a pointer to a int64.
// If n is zero, it returns nil.
func Int64Nil(n int64) *int64 {
	if n == 0 {
		return nil
	}
	return &n
}
