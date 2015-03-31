package gostick

// Buffer is a []byte container with some logic for use
// as a read/write buffer
type buffer []byte

// Cap will resize the buffer slice to full capacity
func (b *buffer) cap() {
	*b = (*b)[:cap(*b)]
}

// New returns the current free space as a new slice
func (b *buffer) new() buffer {
	if len(*b) < cap(*b) {
		return (*b)[len(*b):cap(*b)]
	}
	return nil
}

// Resize changes the slice length with i bytes. It is
// not an error to have a negative i.
func (b *buffer) resize(i int) {
	if i == 0 {
		return
	}
	i = len(*b) + i
	if i < 0 {
		i = 0
	}
	if i > cap(*b) {
		i = cap(*b)
	}
	*b = (*b)[:i]
}

// Shift out i bytes from the slice sliding remaining data down
func (b *buffer) shift(i int) {
	if i <= 0 {
		return
	}
	if i < len(*b) {
		copy(*b, (*b)[i:])
		*b = (*b)[:len(*b)-i]
	} else {
		*b = (*b)[:0]
	}
}

// Trunc sets the slice length to i
func (b *buffer) trunc(i int) {
	if i < 0 {
		i = 0
	}
	if i > cap(*b) {
		i = cap(*b)
	}
	*b = (*b)[:i]
}
