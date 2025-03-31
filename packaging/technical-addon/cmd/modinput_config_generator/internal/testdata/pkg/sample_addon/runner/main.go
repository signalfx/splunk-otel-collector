package main

import (
	"github.com/splunk/otel-technical-addon/internal/modularinput"
	"os"
	"os/exec"
)

func main() {
	config := GetDefaultSampleAddonModularInputs()

	// Create a new transformer with the embedded configuration
	transformer := modularinput.NewModinputTransformer(config.SchemaName, config.ModularInputs)

	// TODO hughesjj
	// validate input (may want to wait until registration)
	// 1. set diff of required - provided, err if nonempty
	// 2. set diff of provided - valid, err if nonempty

	xmlInput, err := modularinput.ReadXML(os.Stdin)
	if err != nil {
		panic(err)
	}
	err = transformer.Transform(xmlInput)
	if err != nil {
		panic(err)
	}
	flags := transformer.GetFlags()
	envVars := transformer.GetEnvVars()

	// TODO hughesjj
	// 1. better logging
	// 2. Shell wrapper & invoke it
	cmd := exec.Command("true", flags...)
	cmd.Env = append(os.Environ(), envVars...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		panic(err)
	}

}
