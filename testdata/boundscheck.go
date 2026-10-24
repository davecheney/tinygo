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
}
