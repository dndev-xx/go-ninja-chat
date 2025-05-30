package pointer

// Indirect returns the value from the passed pointer or the zero value if the pointer is nil.
// Inspired by reflect.Indirect.
func Indirect[T any](ptr *T) T {
	var zero T
	if ptr == nil {
		return zero
	}
	return *ptr
}

// Ptr returns pointer to value v.
func Ptr[T any](v T) *T {
	return &v
}

// PtrWithZeroAsNil returns pointer to value v.
// But if v has zero value then PtrWithZeroAsNil returns nil.
func PtrWithZeroAsNil[T comparable](v T) *T {
	var zero T
	if v == zero {
		return nil
	}
	return &v
}
