package main

// Bounds checks that the compiler can prove, and cases that look similar but
// must still panic.

var data = []byte("  hello world")
var other = []byte("ab")
var start = -1

//go:noinline
func rangeLoop(s []byte) int {
	n := 0
	for i := range s {
		n += int(s[i])
	}
	return n
}

//go:noinline
func countedLoop(s []byte) int {
	n := 0
	for i := 0; i < len(s); i++ {
		n += int(s[i])
	}
	return n
}

//go:noinline
func whileLoop(s string) int {
	i := 0
	for i < len(s) && s[i] == ' ' {
		i++
	}
	return i
}

//go:noinline
func twoLoops(s string) int {
	i := 0
	for i < len(s) && s[i] == ' ' {
		i++
	}
	n := 0
	for i < len(s) {
		if s[i] != ' ' {
			n++
		}
		i++
	}
	return n
}

//go:noinline
func negativeStart(s []byte) int {
	n := 0
	for i := start; i < len(s); i++ {
		n += int(s[i])
	}
	return n
}

//go:noinline
func otherSlice(s, t []byte) int {
	n := 0
	for i := 0; i < len(t); i++ {
		n += int(s[i])
	}
	return n
}

//go:noinline
func signedConvert(s []byte, b uint8) byte {
	i := int(int8(b))
	if i < len(s) {
		return s[i]
	}
	return 0
}

//go:noinline
func shortened(s []byte) int {
	n := 0
	t := s
	for i := 0; i < len(s); i++ {
		t = t[:1]
		n += int(t[i])
	}
	return n
}

//go:noinline
func lessOrEqual(s []byte) int {
	n := 0
	for i := 0; i <= len(s); i++ {
		n += int(s[i])
	}
	return n
}

//go:noinline
func sliceTail(s []byte) []byte {
	for i := 0; i < len(s); i++ {
		if s[i] == 'w' {
			return s[i:]
		}
	}
	return nil
}

//go:noinline
func stringTail(s string) string {
	i := 0
	for i < len(s) && s[i] == ' ' {
		i++
	}
	return s[i:]
}

//go:noinline
func sliceAtEnd(s []byte, i int) []byte {
	if i >= 0 && i <= len(s) {
		return s[i:]
	}
	return nil
}

//go:noinline
func sliceNegative(s []byte) []byte {
	i := start
	if i < len(s) {
		return s[i:]
	}
	return nil
}

//go:noinline
func sliceOther(s, t []byte) []byte {
	i := len(t) - 1
	if i < len(t) {
		return s[i:]
	}
	return nil
}

//go:noinline
func slicePastEnd(s []byte) []byte {
	i := len(s) + 1
	if i <= len(s)+1 {
		return s[i:]
	}
	return nil
}

//go:noinline
func shrinkTail(s []byte) []byte {
	i := 0
	for i < len(s) {
		i++
		s = s[:len(s)/2] // s gets shorter than i
	}
	return s[i:]
}

func mustPanic(name string, f func()) {
	defer func() {
		if recover() != nil {
			println(name, "panicked")
		}
	}()
	f()
	println(name, "did not panic")
}

func main() {
	println("rangeLoop:", rangeLoop(data))
	println("countedLoop:", countedLoop(data))
	println("whileLoop:", whileLoop(string(data)))
	println("twoLoops:", twoLoops(string(data)))
	println("empty:", rangeLoop(nil), countedLoop(nil), whileLoop(""), twoLoops(""))

	mustPanic("negativeStart", func() { negativeStart(data) })
	mustPanic("otherSlice", func() { otherSlice(other, data) })
	mustPanic("signedConvert", func() { signedConvert(data, 200) })
	mustPanic("shortened", func() { shortened(data) })
	mustPanic("lessOrEqual", func() { lessOrEqual(data) })

	println("sliceTail:", string(sliceTail(data)))
	println("stringTail:", stringTail(string(data)))
	println("sliceAtEnd:", len(sliceAtEnd(data, len(data))), string(sliceAtEnd(data, 8)))
	println("empty tails:", len(sliceTail(nil)), len(stringTail("")), len(sliceAtEnd(nil, 0)))

	mustPanic("sliceNegative", func() { sliceNegative(data) })
	mustPanic("sliceOther", func() { sliceOther(other, data) })
	mustPanic("slicePastEnd", func() { slicePastEnd(data) })
	mustPanic("shrinkTail", func() { shrinkTail(data) })
}
