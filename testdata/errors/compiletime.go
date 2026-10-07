package main

//go:compiletime
func missing()

//go:compiletime unexpected
func invalid() {}

func main() {
	missing()
	invalid()
}

// ERROR: # command-line-arguments
// ERROR: compiletime.go:3:1: can only use //go:compiletime on definitions
// ERROR: compiletime.go:6:1: //go:compiletime does not accept parameters
