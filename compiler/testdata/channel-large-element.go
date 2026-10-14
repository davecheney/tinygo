package main

func makeLargeElementChan(size int) chan [32 << 20]byte {
	return make(chan [32 << 20]byte, size)
}
