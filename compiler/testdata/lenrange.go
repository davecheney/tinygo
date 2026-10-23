package main

func sliceArgs(s []int) int {
	return len(s) + cap(s)
}

func stringArg(s string) int {
	return len(s)
}

func plainInt(n int) int {
	return n
}
