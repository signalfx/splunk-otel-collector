// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/splunk/splunk-technical-addon/internal/addonruntime"
	"github.com/splunk/splunk-technical-addon/internal/modularinput"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	os.Exit(run())
}

func run() int {
	config := GetDefaultSplunkTAOtelCollectorModularInputs()

	schemeFlag := flag.Bool("scheme", false, "Print the scheme and exit")
	validateFlag := flag.Bool("validate-arguments", false, "Validate the arguments and exit")
	flag.Parse()
	if *schemeFlag {
		// TODO we've never actually implemented this, let's figure it out in the future
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
		log.Errorln(fmt.Errorf("error reading modinput: %w", err))
	}

	err = mip.ProcessXML(xmlInput)
	if *validateFlag {
		if err != nil {
			return -1
		}
		return 0
	}
	if err != nil {
		log.Errorln(fmt.Errorf("error parsing modinput: %w", err))
		return -1
	}

	if err = Run(mip); err != nil {
		log.Errorln(fmt.Errorf("error running splunk linux autoinstrumentation addon: %w", err))
		return -1
	}
	return 0
}

func Run(mip *modularinput.ModinputProcessor) error {
	modInputs := GetSplunkTAOtelCollectorModularInputs(mip)
	outfile, err := os.Create(modInputs.SplunkOtelLogFile.Value)
	if err != nil {
		return err
	}
	defer outfile.Close()

	flags := mip.GetFlags()
	envVars := mip.GetEnvVars()
	splunkTaPlatformDir, err := addonruntime.GetTaPlatformDir()
	if err != nil {
		return err
	}

	cmd := exec.Command(filepath.Join(splunkTaPlatformDir, otelBinaryName), flags...)
	cmd.Env = append(os.Environ(), envVars...)
	cmd.Stdout = outfile
	cmd.Stderr = outfile
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
