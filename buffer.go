package gostick

// Buffer is a FIFO buffer with some logic attached
type Buffer []byte

// Cap will resize the buffer length to full capacity
func (b *Buffer) Cap() {
	*b = (*b)[:cap(*b)]
}

// New returns the current free space as a new Buffer
func (b *Buffer) New() Buffer {
	if len(*b) < cap(*b) {
		return (*b)[len(*b):cap(*b)]
	}
	return nil
}

// Grow extends the buffer length with i bytes
func (b *Buffer) Grow(i int) {
	if i == 0 || (i+len(*b)) > cap(*b) {
		i = cap(*b) - len(*b)
	}
	*b = (*b)[:len(*b)+i]
}

// Shift will shift out i bytes moving any remaining data
func (b *Buffer) Shift(i int) {
	if i < len(*b) {
		copy(*b, (*b)[i:])
		*b = (*b)[:len(*b)-i]
	} else {
		*b = (*b)[:0]
	}
}

// Trunc sets the buffer length to i
func (b *Buffer) Trunc(i int) {
	if i < len(*b) {
		*b = (*b)[:i]
	}
}
