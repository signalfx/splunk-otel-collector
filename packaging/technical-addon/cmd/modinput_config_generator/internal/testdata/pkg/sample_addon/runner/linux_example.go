//go:build !windows

package main

import (
	"encoding/json"
	"fmt"
	"log"
)

func Example(flags []string, envVars []string) {
	output, err := json.Marshal(ExampleOutput{flags, envVars, "linux"})
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println(output)
	}
	fmt.Println(output)
}
