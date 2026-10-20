package interp

import "testing"

func TestSplitsPointer(t *testing.T) {
	// p and q are encoded pointer words. All words of a pointer are equal.
	const p, q = 1 << 32, 2 << 32
	tests := []struct {
		buf         []uint64
		start, end  uint32
		pointerSize uint32
		want        bool
	}{
		{[]uint64{1, 2, 3, 4}, 1, 3, 4, false},
		{[]uint64{p, p, p, p, p, p, p, p}, 0, 4, 4, false},
		{[]uint64{p, p, p, p, p, p, p, p}, 4, 8, 4, false},
		{[]uint64{p, p, p, p, p, p, p, p}, 0, 8, 4, false},
		{[]uint64{p, p, p, p, p, p, p, p}, 2, 6, 4, true},
		{[]uint64{p, p, p, p, p, p, p, p}, 0, 1, 4, true},
		{[]uint64{p, p, q, q}, 1, 3, 2, true},
		{[]uint64{p, p, q, q}, 2, 4, 2, false},
		// A store replaced part of the first pointer.
		{[]uint64{0, 0, p, p, p, p, p, p}, 2, 6, 4, true},
		{[]uint64{0, 0, p, p, p, p, p, p}, 4, 8, 4, true},
		{[]uint64{0, 0, p, p, p, p, p, p}, 0, 2, 4, false},
		// A plain byte next to a pointer.
		{[]uint64{0, 0, 0, 0, p, p, p, p}, 3, 5, 4, true},
		{[]uint64{0, 0, 0, 0, p, p, p, p}, 3, 8, 4, false},
		// Pointers can start at an odd offset if the pointer alignment is 1.
		{[]uint64{0, p, p, p, p}, 1, 3, 2, false},
		{[]uint64{0, p, p, p, p}, 3, 5, 2, false},
		{[]uint64{0, p, p, p, p}, 2, 4, 2, true},
	}
	for _, tc := range tests {
		got := rawValue{buf: tc.buf}.splitsPointer(tc.start, tc.end, tc.pointerSize)
		if got != tc.want {
			t.Errorf("splitsPointer(%v, %d, %d, %d) = %v, want %v", tc.buf, tc.start, tc.end, tc.pointerSize, got, tc.want)
		}
	}
}
