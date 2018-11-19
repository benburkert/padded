package padded

import (
	"testing"
)

func TestPool(t *testing.T) {
	pool := new(Pool)

	s1 := pool.Make(128, 256, 64)
	assertLenCapPad(t, s1, 128, 256, 64)
	pool.Free(s1)

	s2 := pool.Make(64, 128, 64)
	assertLenCapPad(t, s2, 64, 128, 64)
	if want, got := &s1[0], &s2[0]; want != got {
		t.Error("s1 not reallocated for s2")
	}
	pool.Free(s2)

	s3 := pool.Make(512, 512, 64)
	defer pool.Free(s3)

	assertLenCapPad(t, s3, 512, 512, 64)
	if ptr, got := &s3[0], &s1[0]; ptr == got {
		t.Errorf("s1 reallocated for s3")
	}
}
