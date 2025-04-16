package main

// Returns "new" initialised memory.
// Calls `new()` on the given type, writes the value to the pointer and returns
// the pointer.
func NewPtr[T any](x T) *T {
	ptr := new(T)
	*ptr = x
	return ptr
}
