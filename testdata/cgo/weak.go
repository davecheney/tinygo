package main

// The strong definitions in main.c replace these weak definitions at link time.
// Default visibility because the builder makes hidden symbols internal.

// __attribute__((weak, visibility("default"))) int weakValue = 1;
// __attribute__((weak, visibility("default"))) int weakFunc(void) { return 1; }
import "C"

var (
	weakValueInit = int(C.weakValue)
	weakFuncInit  = int(C.weakFunc())
)

func init() {
	C.weakValue = 7
}

func weakLinkage() {
	println("weak value:", weakValueInit, int(C.weakValue))
	println("weak func:", weakFuncInit)
}
