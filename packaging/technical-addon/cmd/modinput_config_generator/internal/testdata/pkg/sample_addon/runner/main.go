package main

import (
	"flag"
	"fmt"
	"github.com/splunk/otel-technical-addon/internal/modularinput"
	"io"
	"log"
	"os"
	"path/filepath"
)

type ExampleOutput struct {
	Flags    []string
	EnvVars  []string
	Platform string
}

func setupLogger(schemaName string) (closer func(), err error) {
	// Setup the logger prefix to the proper date time format.
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.LUTC | log.Lshortfile)

	logFilePath, ok := buildLogFilePath(schemaName)
	if !ok {
		return func() {}, fmt.Errorf("could not find log file path")
	}

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Error: failed to open log file %s: %v", logFilePath, err)
		return func() {}, fmt.Errorf("could not open log file %s", logFilePath)
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)

	return func() { _ = logFile.Close() }, nil
}

func buildLogFilePath(modularinputName string) (string, bool) {
	splunkHome, ok := os.LookupEnv("SPLUNK_HOME")
	if !ok {
		// Do not log to a file if SPLUNK_HOME is not set.
		return "", false
	}

	return filepath.Join(splunkHome, "var", "log", "splunk", modularinputName+".log"), true
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

	logCloseFunc, err := setupLogger(config.SchemaName)
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
