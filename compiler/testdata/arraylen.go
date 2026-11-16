package main

// The array type in these functions is longer than the largest uintptr on a
// target with 16-bit pointers. Each length constant must be the largest
// uintptr, not a truncated value.

func indexBig(p *[1 << 16]byte, i uintptr) byte {
	return p[i]
}

func sliceBig(p *[1 << 16]byte) []byte {
	return p[:]
}

func sliceToBig(s []byte) *[1 << 16]byte {
	return (*[1 << 16]byte)(s)
}

func boxBig(p *[1 << 16]byte) any {
	return p
}
