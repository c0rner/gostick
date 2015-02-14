package gostick

// Buffer is a []byte container with some logic for use
// as a read/write buffer attached
type buffer []byte

// Cap will resize the buffer length to full capacity
func (b *buffer) cap() {
	*b = (*b)[:cap(*b)]
}

// New returns the current free space as a new buffer
func (b *buffer) new() buffer {
	if len(*b) < cap(*b) {
		return (*b)[len(*b):cap(*b)]
	}
	return nil
}

// Resize changes the buffer length with i bytes
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

// Shift will shift out i bytes moving any remaining data
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

// Trunc sets the buffer length to i
func (b *buffer) trunc(i int) {
	if i < 0 {
		i = 0
	}
	if i > cap(*b) {
		i = cap(*b)
	}
	*b = (*b)[:i]
}
