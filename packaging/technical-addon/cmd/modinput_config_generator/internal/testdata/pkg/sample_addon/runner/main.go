package main

import (
	"github.com/splunk/otel-technical-addon/internal/modularinput"
	"os"
)

func main() {
	config := GetDefaultSampleAddonModularInputs()

	// Create a new modular input processor with the embedded configuration
	mip := modularinput.NewModinputProcessor(config.SchemaName, config.ModularInputs)

	xmlInput, err := modularinput.ReadXML(os.Stdin)
	if err != nil {
		panic(err)
	}
	err = mip.ProcessXml(xmlInput)
	if err != nil {
		panic(err)
	}
	flags := mip.GetFlags()
	envVars := mip.GetEnvVars()
	Example(flags, envVars)
}
