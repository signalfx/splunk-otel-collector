// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"os"

	"github.com/splunk/splunk-technical-addon/internal/modularinput"
)

type ExampleOutput struct {
	Flags   []string
	EnvVars []string

	SplunkHome   string
	TaHome       string
	PlatformHome string

	EverythingSet              string
	MinimalSet                 string
	MinimalSetRequired         string
	UnaryFlagWithEverythingSet string

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
	err = mip.ProcessXML(xmlInput)

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
	Example(mip)
	return 0
}
