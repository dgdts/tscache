package tscache

// ByteView represents an immutable view of bytes.
type ByteView struct {
	B []byte // B is the slice of bytes
}

// Len returns the length of the byte slice.
func (v ByteView) Len() int {
	return len(v.B)
}

// ByteSlice returns a copy of the byte slice.
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.B)
}

// cloneBytes makes a copy of the provided byte slice.
func cloneBytes(b []byte) []byte {
	ret := make([]byte, len(b))
	copy(ret, b)
	return ret
}

// String returns the string representation of the byte slice.
func (v ByteView) String() string {
	return string(v.B) // Convert byte slice to string
}
