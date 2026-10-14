package main

import (
	"crypto/fips140"
	"crypto/sha3"
	"os"
)

func main() {
	if fips140.Enabled() {
		panic("FIPS mode is enabled")
	}
	run()
}

func digest() [32]byte {
	return sha3.Sum256([]byte(os.Args[0]))
}
