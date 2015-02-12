package gostick

type Buffer []byte

func (b *Buffer) Cap() {
	*b = (*b)[:cap(*b)]
}

func (b *Buffer) New() Buffer {
	if len(*b) < cap(*b) {
		return (*b)[len(*b):cap(*b)]
	}
	return nil
}

func (b *Buffer) Grow(i int) {
	if i == 0 || (i+len(*b)) > cap(*b) {
		i = cap(*b) - len(*b)
	}
	*b = (*b)[:len(*b)+i]
}

func (b *Buffer) Shift(i int) {
	if i < len(*b) {
		copy(*b, (*b)[i:])
		*b = (*b)[:len(*b)-i]
	} else {
		*b = (*b)[:0]
	}
}

func (b *Buffer) Trunc(i int) {
	if i < len(*b) {
		*b = (*b)[:i]
	}
}
