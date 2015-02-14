package gostick

import (
	"testing"
)

const (
	hexstring = "0123456789ABCDEF"
)

var (
	buf = make(buffer, 0, len(hexstring))
)

func Test_bufferCap(t *testing.T) {
	buf.trunc(0)
	buf.cap()
	if len(buf) != cap(buf) {
		t.Fail()
	}
}

func Test_bufferResize(t *testing.T) {
	var tests = []struct {
		start  int
		resize int
		result int
	}{
		{start: 0, resize: -1, result: 0},
		{start: 5, resize: -1, result: 4},
		{start: 5, resize: 0, result: 5},
		{start: 5, resize: 1, result: 6},
		{start: 0, resize: cap(buf), result: cap(buf)},
		{start: 5, resize: cap(buf), result: cap(buf)},
	}
	for _, test := range tests {
		buf.trunc(test.start)
		buf.resize(test.resize)
		if len(buf) != test.result {
			t.Errorf("resize(%d): Wanted (%d), got '%d'\n", test.resize, test.result, len(buf))
		}
	}
}

func Test_bufferNew(t *testing.T) {
	var tests = []struct {
		shift  int
		result string
	}{
		{shift: 5, result: "56789ABCDEF01234"},
		{shift: 10, result: "ABCDEF0123456789"},
	}
	buf.cap()
	for _, test := range tests {
		copy(buf, hexstring)
		buf.shift(test.shift)
		newbuf := buf.new()
		copy(newbuf, hexstring)
		buf.cap()
		for i := 0; i < len(buf); i++ {
			if buf[i] != test.result[i] {
				t.Errorf("Wanted '%s' (%d), got '%s' (%d)\n", test.result, len(test.result), buf, len(buf))
				break
			}
		}
	}
}

func Test_bufferShift(t *testing.T) {
	var tests = []struct {
		shift  int
		result string
	}{
		{shift: -1, result: hexstring},
		{shift: 0, result: hexstring},
		{shift: 1, result: "123456789ABCDEF"},
		{shift: cap(buf) - 1, result: "F"},
		{shift: cap(buf), result: ""},
		{shift: cap(buf) + 1, result: ""},
	}

	for _, test := range tests {
		buf.cap()
		copy(buf, hexstring)
		buf.shift(test.shift)
		if len(buf) != len(test.result) {
			t.Errorf("Wanted '%s' (%d), got '%s' (%d)\n", test.result, len(test.result), buf, len(buf))
		}
		for i := 0; i < len(buf); i++ {
			if buf[i] != test.result[i] {
				t.Errorf("Wanted '%s' (%d), got '%s' (%d)\n", test.result, len(test.result), buf, len(buf))
			}
		}
	}
}

func Test_bufferTrunc(t *testing.T) {
	var tests = []struct {
		trunc  int
		result int
	}{
		{trunc: -1, result: 0},
		{trunc: 0, result: 0},
		{trunc: 1, result: 1},
		{trunc: cap(buf) - 1, result: cap(buf) - 1},
		{trunc: cap(buf), result: cap(buf)},
		{trunc: cap(buf) + 1, result: cap(buf)},
	}
	for _, test := range tests {
		buf.trunc(test.trunc)
		if len(buf) != test.result {
			t.Errorf("trunc(%d): Wanted (%d), got '%d'\n", test.trunc, test.result, len(buf))
		}
	}
}
