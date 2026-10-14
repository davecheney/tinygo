//go:build sha3used

package main

func run() {
	sum := digest()
	println(sum[0])
}
