package main

func testRangeString() {
	for i, c := range "abcü¢€𐍈°x" {
		println(i, c)
	}
}

func testStringToRunes() {
	var s = "abcü¢€𐍈°x"
	for i, c := range []rune(s) {
		println(i, c)
	}
}

func testRunesToString(r []rune) {
	println("string from runes:", string(r))
}

func testByteSliceStringOrderedCompare() {
	a := []byte("abc")
	b := []byte("abd")
	println("a < b:", string(a) < string(b))
	println("a <= b:", string(a) <= string(b))
	println("a > b:", string(a) > string(b))
	println("a >= b:", string(a) >= string(b))
	println("a < a:", string(a) < string(a))
	println("a <= a:", string(a) <= string(a))

	// A mutation between the two conversions must still be observed: the
	// comparison must use the value of s after mutateFirstByte runs, not a
	// stale copy of the slice contents.
	c := []byte("abc")
	println("mutated a < c:", string(a) < string(mutateFirstByte(c)))
}

//go:noinline
func mutateFirstByte(s []byte) []byte {
	s[0] = 'z'
	return s
}

type myString string

func main() {
	testRangeString()
	testStringToRunes()
	testRunesToString([]rune{97, 98, 99, 252, 162, 8364, 66376, 176, 120})
	testByteSliceStringOrderedCompare()
	var _ = len([]byte(myString("foobar"))) // issue 1246
}
