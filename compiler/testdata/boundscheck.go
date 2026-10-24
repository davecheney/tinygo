package main

// The index checks in these loops are proven by the compiler and not emitted.

func countedLoop(s []byte) int {
	n := 0
	for i := 0; i < len(s); i++ {
		n += int(s[i])
	}
	return n
}

func whileLoop(s string) int {
	i := 0
	for i < len(s) && s[i] == ' ' {
		i++
	}
	return i
}

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

// The index can be negative, so the check stays.

func negativeStart(s []byte, start int) int {
	n := 0
	for i := start; i < len(s); i++ {
		n += int(s[i])
	}
	return n
}
