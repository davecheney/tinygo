package interp

import "testing"

func TestInsidePointer(t *testing.T) {
	// p and q are encoded pointer words. All words of a pointer are equal.
	const p, q = 1 << 32, 2 << 32
	tests := []struct {
		buf         []uint64
		offset      uint32
		pointerSize uint32
		want        bool
	}{
		{[]uint64{1, 2, 3, 4}, 1, 4, false},
		{[]uint64{p, p, p, p, p, p, p, p}, 0, 4, false},
		{[]uint64{p, p, p, p, p, p, p, p}, 4, 4, false},
		{[]uint64{p, p, p, p, p, p, p, p}, 2, 4, true},
		{[]uint64{p, p, q, q}, 1, 2, true},
		{[]uint64{p, p, q, q}, 2, 2, false},
		// A store replaced part of the first pointer.
		{[]uint64{0, 0, p, p, p, p, p, p}, 2, 4, true},
		{[]uint64{0, 0, p, p, p, p, p, p}, 4, 4, true},
		// Pointers can start at an odd offset if the pointer alignment is 1.
		{[]uint64{0, p, p, p, p}, 1, 2, false},
		{[]uint64{0, p, p, p, p}, 3, 2, false},
		{[]uint64{0, p, p, p, p}, 2, 2, true},
	}
	for _, tc := range tests {
		got := rawValue{buf: tc.buf}.insidePointer(tc.offset, tc.pointerSize)
		if got != tc.want {
			t.Errorf("insidePointer(%v, %d, %d) = %v, want %v", tc.buf, tc.offset, tc.pointerSize, got, tc.want)
		}
	}
}
