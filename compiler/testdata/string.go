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

func byteSliceStringCompareLess(s1, s2 []byte) bool {
	return string(s1) < string(s2)
}

func byteSliceStringCompareLessEqual(s1, s2 []byte) bool {
	return string(s1) <= string(s2)
}

func byteSliceStringCompareGreater(s1, s2 []byte) bool {
	return string(s1) > string(s2)
}

func byteSliceStringCompareGreaterEqual(s1, s2 []byte) bool {
	return string(s1) >= string(s2)
}

func byteSliceStringCompareLessSideEffects(s1, s2 []byte) bool {
	return string(s1) < string(mutateBytes(s2))
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
