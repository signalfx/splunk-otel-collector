//go:build windows

package main

import (
	"encoding/json"
	"log"
)

func Example(flags []string, envVars []string) {
	log.Println(json.Marshal(ExampleOutput{flags, envVars, "windows"}))
}
