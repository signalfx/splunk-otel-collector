//go:build windows

package main

import "fmt"

func Example(flags []string, envVars []string) {
	fmt.Printf("for windows, flags: %v ; env vars: %v", flags, envVars)
}
