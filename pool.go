package padded

import (
	"reflect"
	"sync"
	"unsafe"
)

// Pool implements a memory pool for padded byte slices. Both the capacity and
// padding are reclaimed when a slice is freed.
type Pool struct {
	pool sync.Pool
}

// Make allocates (or reallocates) a Slice for the len, cap, and pad.
func (p *Pool) Make(len, cap, pad int) Slice {
	if pad%wordSize != 0 {
		pad = (1 + (pad / wordSize)) * wordSize
	}

	if blk := p.get(pad + cap + wordSize); blk != nil {
		return realloc(*blk, len, cap, pad)
	}

	return Make(len, cap, pad)
}

// Free returns a Slice to p for future reallocation.
func (p *Pool) Free(s Slice) {
	pad := s.Pad() + wordSize

	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&s))
	hdr.Data -= uintptr(pad)
	hdr.Cap += pad
	hdr.Len = hdr.Cap

	p.put(s)
}

func (p *Pool) get(min int) *[]byte {
	if blk, ok := p.pool.Get().(*[]byte); ok && len(*blk) >= min {
		return blk
	}
	return nil
}

func (p *Pool) put(s []byte) {
	p.pool.Put((*[]byte)(&s))
}

func realloc(blk []byte, len, cap, pad int) Slice {
	pad += wordSize
	initPadding(blk[:pad])
	return blk[pad : pad+len : pad+cap]
}
