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
		log.Fatal("error marshalling json: ", err)
	} else {
		log.Printf("Sample output:%s\n", output)
	}
	fmt.Printf("Sample output:%s\n", output)
}
