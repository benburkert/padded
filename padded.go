// Package padded provides a padded byte slice and pool allocator.
package padded

import (
	"encoding/binary"
	"reflect"
	"unsafe"
)

const wordSize = int(unsafe.Sizeof(int(0)))

// Slice is a padded byte slice. The padding is a scratch space allocated in
// front of byte slice and behaves roughly the same as the capacity. The len &
// cap built-in funcs work the same for Slice as for []byte, but append does
// not. Use the Append method instead.
//
// Prepending elements to the slice fills in the padding, just as appending
// elements to the slice fills in the capacity. The elements must be naturally
// aligned to fill in the scratch padding, otherwise an append is performed.
// The word size on an amd64 machine is 8 bytes, so element bytes should be
// prepended in multiples of 8.
type Slice []byte

// Make allocates and initializes a padded byte slice. Unlike cap, the pad
// value is only the size of the scratch space in front of the slice. (The
// scratch space in back is cap - len.)
//
// The allocated scratch space for padding may be >= pad.
func Make(len, cap, pad int) Slice {
	if pad%wordSize != 0 {
		pad = (1 + (pad / wordSize)) * wordSize
	}
	pad += wordSize

	slice := make([]byte, pad+len, pad+cap)
	initPadding(slice[:pad])
	return slice[pad : pad+len : pad+cap]
}

// Append appends elements to the end of s. The behavior of Append is the same
// as the built-in append function.
func (s Slice) Append(elems ...byte) Slice {
	if cap(s) < len(s)+len(elems) {
		slice := append(make([]byte, wordSize), append(s, elems...)...)
		return slice[wordSize:]
	}
	return append(s, elems...)
}

// Pad returns the length of the remaining bytes of padding.
func (s Slice) Pad() int {
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&s))
	ptr := hdr.Data - uintptr(wordSize)
	buf := (*[wordSize]byte)(unsafe.Pointer(ptr))

	pad, _ := binary.Uvarint(buf[:])
	return int(pad) * wordSize
}

// Prepend concatenates elems onto the beginning of s.
func (s Slice) Prepend(elems ...byte) Slice {
	pad := len(elems)
	if pad%wordSize != 0 || pad > s.Pad() {
		return append(make([]byte, wordSize), append(elems, s...)...)[wordSize:]
	}

	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&s))
	hdr.Data -= uintptr(pad)
	hdr.Len += pad
	hdr.Cap += pad

	copy(s, elems)
	return s
}

func initPadding(s []byte) {
	for i := 0; i < len(s)/wordSize; i++ {
		idx := i * wordSize
		binary.PutUvarint(s[idx:idx+wordSize], uint64(i))
	}
}
