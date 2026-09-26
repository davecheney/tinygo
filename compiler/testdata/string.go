package main

func someString() string {
	return "foo"
}

func zeroLengthString() string {
	return ""
}

func stringLen(s string) int {
	return len(s)
}

func stringIndex(s string, index int) byte {
	return s[index]
}

func stringCompareEqual(s1, s2 string) bool {
	return s1 == s2
}

func stringCompareUnequal(s1, s2 string) bool {
	return s1 != s2
}

func byteSliceStringCompareEqual(s1, s2 []byte) bool {
	return string(s1) == string(s2)
}

func byteSliceStringCompareUnequal(s1, s2 []byte) bool {
	return string(s1) != string(s2)
}

func byteSliceStringCompareSideEffects(s1, s2 []byte) bool {
	return string(s1) == string(mutateBytes(s2))
}

func byteSliceStringCompareNil(s []byte) bool {
	var nilSlice []byte
	return string(s) == string(nilSlice)
}

func byteSliceStringCompareLocal(a, b []byte) bool {
	s := string(a)
	t := string(b)
	return s == t
}

func byteSliceStringCompareEscape(a, b []byte) (bool, string) {
	s := string(a)
	t := string(b)
	return s == t, s
}

func byteSliceStringCompareStore(a, b []byte, dst *string) bool {
	s := string(a)
	t := string(b)
	equal := s == t
	*dst = s
	return equal
}

func byteSliceStringCompareBox(a, b []byte) (bool, any) {
	s := string(a)
	t := string(b)
	return s == t, any(s)
}

func byteSliceStringCompareMutation(a, b []byte) bool {
	s := string(a)
	mutateBytes(b)
	return s == string(b)
}

func byteSliceStringCompareAfterMutation(a, b []byte) (bool, string) {
	s := string(a)
	t := string(b)
	equal := s == t
	mutateBytes(a)
	return equal, s
}

func byteSliceStringCompareSlices(a, b []byte) bool {
	return string(a[:2]) == string(b[:2])
}

func byteSliceStringCompareLiteral(a []byte) bool {
	return string(a) == "abc"
}

func byteSliceStringCompareLess(a, b []byte) bool {
	return string(a) < string(b)
}

func byteSliceStringCompareLessEqual(a, b []byte) bool {
	return string(a) <= string(b)
}

func byteSliceStringCompareGreater(a, b []byte) bool {
	return string(a) > string(b)
}

func byteSliceStringCompareGreaterEqual(a, b []byte) bool {
	return string(a) >= string(b)
}

func byteSliceStringLessEscape(a, b []byte) (bool, string) {
	s := string(a)
	t := string(b)
	return s < t, s
}

func byteSliceStringLessEqualEscape(a, b []byte) (bool, string) {
	s := string(a)
	t := string(b)
	return s <= t, s
}

func byteSliceStringGreaterEscape(a, b []byte) (bool, string) {
	s := string(a)
	t := string(b)
	return s > t, s
}

func byteSliceStringGreaterEqualEscape(a, b []byte) (bool, string) {
	s := string(a)
	t := string(b)
	return s >= t, s
}

func byteSliceStringLessMutation(a, b []byte, c byte) bool {
	s := string(a)
	b[0] = c
	return s < string(b)
}

func byteSliceStringLessEqualMutation(a, b []byte, c byte) bool {
	s := string(a)
	b[0] = c
	return s <= string(b)
}

func byteSliceStringGreaterMutation(a, b []byte, c byte) bool {
	s := string(a)
	b[0] = c
	return s > string(b)
}

func byteSliceStringGreaterEqualMutation(a, b []byte, c byte) bool {
	s := string(a)
	b[0] = c
	return s >= string(b)
}

func byteSliceStringOrderReuse(a, b []byte) (bool, bool, bool, bool) {
	s := string(a)
	t := string(b)
	return s < t, s <= t, s > t, s >= t
}

func byteSliceStringOrderLiteral(a []byte) bool {
	return string(a) < "abc"
}

func byteSliceStringOrderSlices(a, b []byte) bool {
	return string(a[:2]) >= string(b[:2])
}

type namedByte byte
type namedBytes []namedByte
type namedRune rune
type namedString string

func namedByteSliceStringCompare(a, b namedBytes) bool {
	return string(a) == string(b)
}

func byteSliceNamedStringCompare(a, b []byte) bool {
	return namedString(a) != namedString(b)
}

func namedByteSliceStringOrder(a, b namedBytes) (bool, bool, bool, bool) {
	return string(a) < string(b), namedString(a) <= namedString(b),
		string(a) > string(b), namedString(a) >= namedString(b)
}

func namedByteSliceToString(a namedBytes) string {
	return string(a)
}

func namedByteSliceToNamedString(a namedBytes) namedString {
	return namedString(a)
}

func namedRuneSliceToString(a []namedRune) string {
	return string(a)
}

//go:noinline
func mutateBytes(s []byte) []byte {
	s[0]++
	return s
}

func stringCompareLarger(s1, s2 string) bool {
	return s1 > s2
}

func stringLookup(s string, x uint8) byte {
	// Test that x is correctly extended to an uint before comparison.
	return s[x]
}
