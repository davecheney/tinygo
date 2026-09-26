package main

import "bytes"

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

func testByteSliceStringCompareNil() {
	var nilSlice []byte
	empty := []byte{}
	full := []byte("foo")
	println(string(nilSlice) == string(empty))
	println(string(nilSlice) == string(full))
	println(string(empty) == string(nilSlice))
	println(string(nilSlice) != string(empty))
}

type myString string

type myByte byte
type myBytes []myByte
type myRune rune

//go:noinline
func compareByteStrings(a, b []byte) (bool, bool) {
	return string(a) == string(b), string(a) != string(b)
}

//go:noinline
func compareLocalByteStrings(a, b []byte) bool {
	s := string(a)
	t := string(b)
	return s == t
}

//go:noinline
func escapeByteString(a, b []byte) (bool, string) {
	s := string(a)
	t := string(b)
	return s == t, s
}

//go:noinline
func storeByteString(a, b []byte, dst *string) bool {
	s := string(a)
	t := string(b)
	equal := s == t
	*dst = s
	return equal
}

//go:noinline
func boxByteString(a, b []byte) (bool, any) {
	s := string(a)
	t := string(b)
	return s == t, any(s)
}

//go:noinline
func mutateByteString(a []byte) (bool, string) {
	s := string(a)
	a[0] = 'z'
	return s == string(a), s
}

//go:noinline
func compareNamedByteStrings(a, b myBytes) (bool, bool) {
	return string(a) == string(b), myString(a) != myString(b)
}

//go:noinline
func convertNamedBytes(a myBytes) (string, myString) {
	return string(a), myString(a)
}

//go:noinline
func convertNamedRunes(a []myRune) string {
	return string(a)
}

//go:noinline
func compareByteStringFallbacks(a, b []byte, s string) {
	println("fallbacks:", string(a) == s, string(a) == "abc",
		string(a[:2]) == string(b[:2]), string(a) < s)
	var x, y any = string(a), string(b)
	a[0] = 'z'
	println("boxed:", x == y, x.(string), y.(string))
}

//go:noinline
func findBytes(a, b []byte) int {
	return bytes.Index(a, b)
}

func testByteSliceStringComparisons() {
	for _, tc := range []struct{ a, b []byte }{
		{nil, nil},
		{nil, []byte{}},
		{[]byte{}, nil},
		{nil, []byte{'a'}},
		{[]byte{'a'}, []byte{'a'}},
		{[]byte{'a'}, []byte{'a', 'a'}},
		{[]byte{'a', 'b'}, []byte{'a', 'c'}},
		{[]byte{0, 0xff}, []byte{0, 0xff}},
	} {
		equal, unequal := compareByteStrings(tc.a, tc.b)
		println("compare:", equal, unequal, compareLocalByteStrings(tc.a, tc.b))
	}

	a := []byte("abc")
	b := []byte("abc")
	equal, snapshot := escapeByteString(a, b)
	a[0] = 'z'
	println("escape:", equal, snapshot, string(a))
	a[0] = 'a'
	equal = storeByteString(a, b, &snapshot)
	a[0] = 'z'
	println("store:", equal, snapshot, string(a))
	a[0] = 'a'
	equal, boxed := boxByteString(a, b)
	a[0] = 'z'
	println("box:", equal, boxed.(string), string(a))
	a[0] = 'a'
	equal, snapshot = mutateByteString(a)
	println("mutation:", equal, snapshot, string(a))
	a[0] = 'a'
	equal, unequal := compareByteStrings(a, b)
	a[0] = 'z'
	println("after:", equal, unequal, string(a))
	a[0] = 'a'
	compareByteStringFallbacks(a, b, "abc")

	named := myBytes{'a', 'b', 'c'}
	equal, unequal = compareNamedByteStrings(named, myBytes{'a', 'b', 'c'})
	s, t := convertNamedBytes(named)
	named[0] = 'z'
	println("named:", equal, unequal, s, t)
	runes := []myRune{'a', 0x20ac, 0x1f600}
	s = convertNamedRunes(runes)
	runes[0] = 'z'
	println("named runes:", s)

	data := []byte("prefix-0123456789abcdefghijkl-suffix")
	sep := []byte("0123456789abcdefghijkl")
	println("index:", findBytes(data, sep), findBytes(sep, data))
	sep[0] = 'z'
	println("index absent:", findBytes(data, sep))
}

//go:noinline
func lessByteStrings(a, b []byte) bool {
	return string(a) < string(b)
}

//go:noinline
func lessEqualByteStrings(a, b []byte) bool {
	return string(a) <= string(b)
}

//go:noinline
func greaterByteStrings(a, b []byte) bool {
	return string(a) > string(b)
}

//go:noinline
func greaterEqualByteStrings(a, b []byte) bool {
	return string(a) >= string(b)
}

//go:noinline
func escapeLessByteString(a, b []byte) (bool, string) {
	s := string(a)
	t := string(b)
	return s < t, s
}

//go:noinline
func escapeLessEqualByteString(a, b []byte) (bool, string) {
	s := string(a)
	t := string(b)
	return s <= t, s
}

//go:noinline
func escapeGreaterByteString(a, b []byte) (bool, string) {
	s := string(a)
	t := string(b)
	return s > t, s
}

//go:noinline
func escapeGreaterEqualByteString(a, b []byte) (bool, string) {
	s := string(a)
	t := string(b)
	return s >= t, s
}

//go:noinline
func mutateLessByteString(a, b []byte, c byte) bool {
	s := string(a)
	b[0] = c
	return s < string(b)
}

//go:noinline
func mutateLessEqualByteString(a, b []byte, c byte) bool {
	s := string(a)
	b[0] = c
	return s <= string(b)
}

//go:noinline
func mutateGreaterByteString(a, b []byte, c byte) bool {
	s := string(a)
	b[0] = c
	return s > string(b)
}

//go:noinline
func mutateGreaterEqualByteString(a, b []byte, c byte) bool {
	s := string(a)
	b[0] = c
	return s >= string(b)
}

//go:noinline
func reuseOrderedByteStrings(a, b []byte) (bool, bool, bool, bool) {
	s := string(a)
	t := string(b)
	return s < t, s <= t, s > t, s >= t
}

//go:noinline
func orderNamedByteStrings(a, b myBytes) (bool, bool, bool, bool) {
	return string(a) < string(b), myString(a) <= myString(b),
		string(a) > string(b), myString(a) >= myString(b)
}

func testByteSliceStringOrdering() {
	for i, tc := range []struct{ a, b []byte }{
		{nil, nil},
		{nil, []byte{}},
		{[]byte{}, nil},
		{nil, []byte{0}},
		{[]byte{0}, nil},
		{[]byte("abc"), []byte("abc")},
		{[]byte("a"), []byte("aa")},
		{[]byte("aa"), []byte("a")},
		{[]byte("abc"), []byte("abd")},
		{[]byte("abc"), []byte("abb")},
		{[]byte{0, 'b'}, []byte{0, 'a'}},
		{[]byte{0, 0x80}, []byte{0, 0xff}},
		{[]byte{0, 0xff}, []byte{0, 0x80}},
		{[]byte{0x7f}, []byte{0x80}},
		{[]byte{0x80}, []byte{0x7f}},
		{[]byte{0xff}, []byte{0xff}},
	} {
		println("order:", i, lessByteStrings(tc.a, tc.b), lessEqualByteStrings(tc.a, tc.b),
			greaterByteStrings(tc.a, tc.b), greaterEqualByteStrings(tc.a, tc.b))
		lt, le, gt, ge := reuseOrderedByteStrings(tc.a, tc.b)
		println("order reuse:", i, lt, le, gt, ge)
	}
	for _, tc := range []struct {
		name   string
		escape func([]byte, []byte) (bool, string)
		mutate func([]byte, []byte, byte) bool
	}{
		{"<", escapeLessByteString, mutateLessByteString},
		{"<=", escapeLessEqualByteString, mutateLessEqualByteString},
		{">", escapeGreaterByteString, mutateGreaterByteString},
		{">=", escapeGreaterEqualByteString, mutateGreaterEqualByteString},
	} {
		for _, c := range []byte{'a', 'm', 'z'} {
			a := []byte("mbc")
			b := []byte{c, 'b', 'c'}
			result, snapshot := tc.escape(a, b)
			a[0] = 'x'
			println("order escape:", tc.name, c, result, snapshot, string(a))
			a[0] = 'm'
			result = tc.mutate(a, a, c)
			println("order mutation:", tc.name, c, result, string(a))
		}
	}
	for _, tc := range []struct{ a, b myBytes }{
		{nil, myBytes{}},
		{myBytes{0, 0x80}, myBytes{0, 0xff}},
		{myBytes{0, 0xff}, myBytes{0, 0x80}},
		{myBytes{'a'}, myBytes{'a', 0}},
		{myBytes{'a', 0}, myBytes{'a'}},
	} {
		lt, le, gt, ge := orderNamedByteStrings(tc.a, tc.b)
		println("named order:", lt, le, gt, ge)
	}
}

func main() {
	testRangeString()
	testStringToRunes()
	testRunesToString([]rune{97, 98, 99, 252, 162, 8364, 66376, 176, 120})
	testByteSliceStringCompareNil()
	testByteSliceStringComparisons()
	testByteSliceStringOrdering()
	var _ = len([]byte(myString("foobar"))) // issue 1246
}
