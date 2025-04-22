//go:build windows

package main

import "fmt"

func Example(flags []string, envVars []string) {
	fmt.Println(json.Marshal(ExampleOutput{flags, envVars, "windows"}))
}
