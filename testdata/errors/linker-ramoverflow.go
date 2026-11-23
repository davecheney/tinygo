package main

var b [256 << 10]byte // 256kB

func main() {
	println("ptr:", &b[0])
}

// ERROR: program uses too much static RAM on this chip (RAM overflowed by {{[0-9]+}} bytes)
