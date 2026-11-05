package main

type sliceField struct {
	data []byte
	n    int
}

type integerFields struct {
	len int
	cap int
}

var globalSlice []byte

var globalString string

func fieldLength(p *sliceField) int {
	s := p.data
	return len(s) + cap(s)
}

func globalLength() int {
	s := globalSlice
	return len(s) + cap(s) + len(globalString)
}

func derefLength(p *[]byte) int {
	return len(*p)
}

func derefString(p *string) int {
	return len(*p)
}

func elementLength(a *[4][]byte, i int) int {
	return len(a[i])
}

func ordinaryFields(p *integerFields) int {
	return p.len + p.cap
}

func negativeLoadedLength(p *[]byte) bool {
	return len(*p) < 0
}

func negativeLoadedCapacity(p *sliceField) bool {
	return cap(p.data) < 0
}

func negativeLoadedString(p *string) bool {
	return len(*p) < 0
}

func negativeLoadedInt(p *integerFields) bool {
	return p.len < 0
}

func globalSum() int {
	s := globalSlice
	sum := 0
	for i := 0; i < len(s); i++ {
		sum += int(s[i])
	}
	return sum
}
