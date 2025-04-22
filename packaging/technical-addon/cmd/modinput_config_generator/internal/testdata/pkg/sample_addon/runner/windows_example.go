//go:build windows

package main

func Example(flags []string, envVars []string) {
	log.Println(json.Marshal(ExampleOutput{flags, envVars, "windows"}))
}
