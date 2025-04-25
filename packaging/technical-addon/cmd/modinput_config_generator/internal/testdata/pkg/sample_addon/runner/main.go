package main

import (
	"flag"
	"github.com/splunk/otel-technical-addon/internal/modularinput"
	"os"
)

type ExampleOutput struct {
	Flags    []string
	EnvVars  []string
	Platform string
}

func main() {
	os.Exit(run())
}

func run() int {
	config := GetDefaultSampleAddonModularInputs()

	schemeFlag := flag.Bool("scheme", false, "Print the scheme and exit")
	validateFlag := flag.Bool("validate-arguments", false, "Validate the arguments and exit")
	flag.Parse()
	if *schemeFlag {
		return 0
	}
	defaultLogFilepath, err := modularinput.GetDefaultLogFilePath(config.SchemaName)
	if err != nil {
		panic(err)
	}
	logCloseFunc, err := modularinput.SetupAddonLogger(defaultLogFilepath)
	if err != nil {
		panic(err)
	}
	defer logCloseFunc()

	// Create a new modular input processor with the embedded configuration
	mip := modularinput.NewModinputProcessor(config.SchemaName, config.ModularInputs)

	xmlInput, err := modularinput.ReadXML(os.Stdin)
	if err != nil {
		panic(err)
	}
	err = mip.ProcessXml(xmlInput)

	if *validateFlag {
		if err != nil {
			return -1
		} else {
			return 0
		}
	}
	if err != nil {
		panic(err)
	}
	flags := mip.GetFlags()
	envVars := mip.GetEnvVars()
	Example(flags, envVars)
	return 0
}
