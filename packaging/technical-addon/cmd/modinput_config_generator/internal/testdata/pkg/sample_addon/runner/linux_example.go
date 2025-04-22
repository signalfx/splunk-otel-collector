//go:build !windows

package main

import (
	"encoding/json"
	"fmt"
)

func Example(flags []string, envVars []string) {
	fmt.Println(json.Marshal(ExampleOutput{flags, envVars, "linux"}))
}
