package joystick

import "testing"

func TestFit(t *testing.T) {
	for _, value := range []int{-32767, -1, 0, 1, 2, 3, 4, 16384, 32767} {
		if got := fit(value, -32767, 32767, -32767, 32767); got != int16(value) {
			t.Errorf("fit(%d) = %d, want %d", value, got, value)
		}
	}
}

func TestFitRanges(t *testing.T) {
	tests := []struct {
		value          int
		minIn, maxIn   int
		minOut, maxOut int16
		want           int16
	}{
		{-32768, -32768, 32767, -32768, 32767, -32768},
		{0, -32768, 32767, -32768, 32767, 0},
		{32767, -32768, 32767, -32768, 32767, 32767},
		{-32767, -32767, 32767, 32767, -32767, 32767},
		{0, -32767, 32767, 32767, -32767, 0},
		{32767, -32767, 32767, 32767, -32767, -32767},
		{0, 0, 1023, -32767, 32767, -32767},
		{512, 0, 1023, -32767, 32767, 32},
		{1023, 0, 1023, -32767, 32767, 32767},
	}
	for _, tt := range tests {
		if got := fit(tt.value, tt.minIn, tt.maxIn, tt.minOut, tt.maxOut); got != tt.want {
			t.Errorf("fit(%d, %d, %d, %d, %d) = %d, want %d", tt.value, tt.minIn, tt.maxIn, tt.minOut, tt.maxOut, got, tt.want)
		}
	}
}

func TestFitRounding(t *testing.T) {
	for _, ranges := range [][4]int{
		{-32768, 32767, -32768, 32767},
		{0, 65535, -32768, 32767},
		{0, 65535, 32767, -32768},
		{0, 1023, -32767, 32767},
		{1023, 0, -32767, 32767},
		{0, 7, -3, 4},
		{0, 6, 4, -3},
	} {
		minIn, maxIn, minOut, maxOut := ranges[0], ranges[1], ranges[2], ranges[3]
		low, high := minIn, maxIn
		if low > high {
			low, high = high, low
		}
		for value := low; value <= high; value++ {
			want := int16((int64(value)-int64(minIn))*(int64(maxOut)-int64(minOut))/(int64(maxIn)-int64(minIn)) + int64(minOut))
			if got := fit(value, minIn, maxIn, int16(minOut), int16(maxOut)); got != want {
				t.Fatalf("fit(%d, %d, %d, %d, %d) = %d, want %d", value, minIn, maxIn, minOut, maxOut, got, want)
			}
		}
	}
}
