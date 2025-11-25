// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/xml"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/splunk/splunk_otel_dotnet_deployer/internal/modularinput"
)

const (
	modularinputName = "splunk_otel_dotnet_deployer"
)

func main() {
	os.Exit(run())
}

func run() int {
	logCloseFunc := setupLogger()
	defer logCloseFunc()

	schemeFlag := flag.Bool("scheme", false, "Print the scheme and exit")
	validateFlag := flag.Bool("validate-arguments", false, "Validate the arguments and exit")
	flag.Parse()

	if *schemeFlag {
		return 0
	}

	if *validateFlag {
		return 0
	}

	input, err := modularinput.ReadXML(os.Stdin)
	if err != nil {
		log.Println("Error:", err)
		return 1
	}

	prettyInput, err := xml.MarshalIndent(input, "", "  ")
	if err != nil {
		log.Println("Error:", err)
		return 1
	}

	log.Printf("Input:\n%s\n", prettyInput)

	if err := runDeployer(input, os.Stdin, os.Stdout, os.Stderr); err != nil {
		log.Println("Error:", err)
		return 1
	}

	return 0
}

// setupStandardLogger intializes the standard logger with settings appropriate for the deployer.
func setupLogger() (closer func()) {
	// Setup the logger prefix to the proper date time format.
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.LUTC | log.Lshortfile)

	logFilePath, ok := buildLogFilePath()
	if !ok {
		return func() {}
	}

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		log.Printf("Error: failed to open log file %s: %v", logFilePath, err)
		return func() {}
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)

	return func() { _ = logFile.Close() }
}

// buildLogFilePath returns the path to the log file.
func buildLogFilePath() (string, bool) {
	splunkHome, ok := os.LookupEnv("SPLUNK_HOME")
	if !ok {
		// Do not log to a file if SPLUNK_HOME is not set.
		return "", false
	}

	return filepath.Join(splunkHome, "var", "log", "splunk", modularinputName+".log"), true
}
