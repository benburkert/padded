package padded

import (
	"testing"
	"unsafe"
)

func TestPaddedSlice(t *testing.T) {
	t.Run("aligned", func(t *testing.T) {
		s := Make(128, 256, 512)
		assertLenCapPad(t, s, 128, 256, 512)

		// prepend with excess padding resizes
		s = s.Prepend(make([]byte, 448)...)
		assertLenCapPad(t, s, 576, 704, 64)

		// prepend without enough padding reallocs
		s1 := s.Prepend(make([]byte, 96)...)
		assertLenCapPad(t, s1, 672, cap(s1), 0)
		if ptr, got := uintptr(unsafe.Pointer(&s[0])), uintptr(unsafe.Pointer(&s1[0])); ptr == got {
			t.Error("backing array reused in prepend without sufficient padding")
		}

		// prepend with exact padding resizes
		s = s.Prepend(make([]byte, 64)...)
		assertLenCapPad(t, s, 640, 768, 0)

		// append with excess capacity resizes
		s = s.Append(make([]byte, 16)...)
		assertLenCapPad(t, s, 656, 768, 0)

		// append with exact capacity resizes
		s = s.Append(make([]byte, 112)...)
		assertLenCapPad(t, s, 768, 768, 0)

		// append without capacity reallocs
		s2 := s.Append(0)
		assertLenCapPad(t, s2, 769, cap(s2), 0)
		if ptr, got := uintptr(unsafe.Pointer(&s[0])), uintptr(unsafe.Pointer(&s2[0])); ptr == got {
			t.Error("backing array reused in append without sufficient capacity")
		}
	})

	t.Run("unaligned", func(t *testing.T) {
		s := Make(100, 200, 301)
		assertLenCapPad(t, s, 100, 200, 304)

		// unaligned prepend reallocs
		s2 := s.Prepend(0)
		assertLenCapPad(t, s2, 101, cap(s2), 0)
		if ptr, got := uintptr(unsafe.Pointer(&s[0])), uintptr(unsafe.Pointer(&s2[0])); ptr == got {
			t.Error("backing array reused in unaligned prepend")
		}

		// unaligned append resizes
		s = s.Append(0)
		assertLenCapPad(t, s, 101, 200, 304)
	})
}

func assertLenCapPad(t *testing.T, s Slice, l, c, p int) {
	t.Helper()

	if want, got := l, len(s); want != got {
		t.Errorf("want len of %d, got %d", want, got)
	}
	if want, got := c, cap(s); want != got {
		t.Errorf("want cap of %d, got %d", want, got)
	}
	if want, got := p, s.Pad(); want != got {
		t.Errorf("want pad of %d, got %d", want, got)
	}
}

func BenchmarkPadded(b *testing.B) {
	b.Run("small", func(b *testing.B) { benchPadded(b, 32) })
	b.Run("medium", func(b *testing.B) { benchPadded(b, 1024) })
	b.Run("large", func(b *testing.B) { benchPadded(b, 1<<20) })
}

func BenchmarkBuiltIn(b *testing.B) {
	b.Run("small", func(b *testing.B) { benchBuiltIn(b, 32) })
	b.Run("medium", func(b *testing.B) { benchBuiltIn(b, 1024) })
	b.Run("large", func(b *testing.B) { benchBuiltIn(b, 1<<20) })
}

func benchPadded(b *testing.B, chunkSize int) {
	b.Run("prepend", func(b *testing.B) {
		buf := Make(0, chunkSize, chunkSize*b.N)

		for i := 0; i < b.N; i++ {
			buf = buf.Prepend(make([]byte, chunkSize)...)
		}
	})

	b.Run("append", func(b *testing.B) {
		buf := Make(0, chunkSize*b.N, 0)

		for i := 0; i < b.N; i++ {
			buf = buf.Append(make([]byte, chunkSize)...)
		}
	})
}

func benchBuiltIn(b *testing.B, chunkSize int) {
	b.Run("prepend", func(b *testing.B) {
		buf := make([]byte, 0, chunkSize*b.N)

		for i := 0; i < b.N; i++ {
			buf = append(make([]byte, chunkSize), buf...)
		}
	})

	b.Run("append", func(b *testing.B) {
		buf := make([]byte, 0, chunkSize*b.N)

		for i := 0; i < b.N; i++ {
			buf = append(buf, make([]byte, chunkSize)...)
		}
	})
}
