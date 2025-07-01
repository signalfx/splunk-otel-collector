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
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/pkg/fileutils"

	log "github.com/sirupsen/logrus"
	"github.com/splunk/splunk-technical-addon/internal/modularinput"
)

func main() {
	os.Exit(run())
}

func run() int {
	config := GetDefaultSplunkTAOtelLinuxAutoinstrumentationModularInputs()

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
		panic(err)
	}
	err = mip.ProcessXML(xmlInput)

	if *validateFlag {
		if err != nil {
			return -1
		}
		return 0
	}
	if err != nil {
		log.Errorf("Error parsing modinput: %+v", err)
		panic(err)
	}
	modInputs := GetSplunkTAOtelLinuxAutoinstrumentationModularInputs(mip)
	err = Run(modInputs)
	if err != nil {
		log.Errorf("error running splunk linux autoinstrumentation addon: %+v", err)
		panic(err)
	}
	// TODO set up traps, but not urgent given this is a run-once style script
	return 0
}

func Run(modInputs *SplunkTAOtelLinuxAutoinstrumentationModularInputs) error {
	lowerModInput := strings.ToLower(modInputs.Remove.Value)
	if "false" == lowerModInput {
		return Instrument(modInputs)
	}
	if "true" == lowerModInput {
		return DeInstrument(modInputs)
	}

	return fmt.Errorf("unknown value for 'remove' modular input, expected (true|false) given %v", strings.ToLower(modInputs.Remove.Value))
}

func DeInstrument(modInputs *SplunkTAOtelLinuxAutoinstrumentationModularInputs) error {
	if err := RemoveJavaInstrumentation(modInputs); err != nil {
		return err
	}
	if err := RemovePreloadInstrumentation(modInputs); err != nil {
		return err
	}
	return nil
}

func Instrument(modInputs *SplunkTAOtelLinuxAutoinstrumentationModularInputs) error {
	if err := InstrumentJava(modInputs); err != nil {
		return err
	}
	if err := AutoinstrumentLdPreload(modInputs); err != nil {
		return err
	}
	return nil
}

func AutoinstrumentLdPreload(modInputs *SplunkTAOtelLinuxAutoinstrumentationModularInputs) error {
	found, err := grepFile(modInputs.AutoinstrumentationPath.Value, modInputs.AutoinstrumentationPreloadPath.Value)
	if err != nil {
		return err
	}
	if !found {
		if strings.ToLower(modInputs.Backup.Value) != "false" {
			if err = backupFile(modInputs.AutoinstrumentationPreloadPath.Value); err != nil {
				return err
			}
		}
		f, err2 := os.OpenFile(modInputs.AutoinstrumentationPreloadPath.Value, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err2 != nil {
			return fmt.Errorf("error opening %s: %w", modInputs.AutoinstrumentationPreloadPath.Value, err2)
		}
		defer f.Close()
		if _, err2 = f.WriteString(modInputs.AutoinstrumentationPath.Value + "\n"); err2 != nil {
			return fmt.Errorf("error writing to %s: %w", modInputs.AutoinstrumentationPreloadPath.Value, err2)
		}
		log.Printf("Successfully autoinstrumented preload at %v with %v\n", modInputs.AutoinstrumentationPreloadPath.Value, modInputs.AutoinstrumentationPath.Value)
	} else {
		log.Printf("Preload already autoinstrumented with %v at %v\n", modInputs.AutoinstrumentationPath.Value, modInputs.AutoinstrumentationPreloadPath.Value)
	}
	return nil
}

func RemovePreloadInstrumentation(modInputs *SplunkTAOtelLinuxAutoinstrumentationModularInputs) error {
	found, err := grepFile(modInputs.Remove.Value, modInputs.AutoinstrumentationPreloadPath.Value)
	if err != nil {
		return err
	}
	if found {
		content, err := os.ReadFile(modInputs.AutoinstrumentationPreloadPath.Value)
		if err != nil {
			return err
		}
		if strings.ToLower(modInputs.Backup.Value) != "false" {
			if err = backupFile(modInputs.AutoinstrumentationPreloadPath.Value); err != nil {
				return err
			}
		}
		newContent := strings.ReplaceAll(string(content), modInputs.AutoinstrumentationPreloadPath.Value, "")
		return os.WriteFile(modInputs.AutoinstrumentationPreloadPath.Value, []byte(newContent), 0644) // #nosec G306

	}
	log.Printf("Autoinstrumentation preload (%s) does not exist or does not contain configured autoinstrumentation (%s)", modInputs.AutoinstrumentationPreloadPath.Value, modInputs.AutoinstrumentationPath.Value)
	return nil
}

func backupFile(currPath string) error {
	if _, err := os.Stat(currPath); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	dir := filepath.Dir(currPath)
	ogFilename := filepath.Base(currPath)
	newPathName := filepath.Join(dir, fmt.Sprintf("%s.%v", ogFilename, time.Now().UnixNano()))
	if _, err := fileutils.CopyFile(currPath, newPathName); err != nil {
		return err
	}
	return nil
}

func grepFile(search string, filepath string) (bool, error) {
	file, err := os.Open(filepath)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), search) {
			return true, nil
		}
	}
	return false, nil
}
