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
	modInputs := GetSplunkTAOtelLinuxAutoinstrumentationModularInputs(mip)
	if err = Run(modInputs); err != nil {
		log.Errorln(fmt.Errorf("error running splunk linux autoinstrumentation addon: %w", err))
		return -1
	}
	// TODO set up traps for completeness, but not urgent given this is a run-once style script
	return 0
}

func Run(modInputs *SplunkTAOtelLinuxAutoinstrumentationModularInputs) error {
	shouldRemove := strings.ToLower(modInputs.Remove.Value)
	if shouldRemove == "false" {
		return Instrument(modInputs)
	}
	if shouldRemove == "true" {
		return DeInstrument(modInputs)
	}

	return fmt.Errorf("unknown value for 'remove' modular input, expected (true|false) given %q", shouldRemove)
}

func DeInstrument(modInputs *SplunkTAOtelLinuxAutoinstrumentationModularInputs) error {
	if err := RemovePreloadInstrumentation(modInputs); err != nil {
		return err
	}
	if err := RemoveJavaInstrumentation(modInputs); err != nil {
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
	if !strings.EqualFold(modInputs.AutoinstrumentationEnabled.Value, "true") {
		log.Println("Autoinstrumentation of /etc/ld.so.preload is disabled in inputs.conf, not configuring")
		return nil
	}
	found, err := stringContainedInFile(modInputs.AutoinstrumentationPath.Value, modInputs.AutoinstrumentationPreloadPath.Value)
	if err != nil {
		return err
	}
	if found {
		log.Printf("Preload already autoinstrumented with %q at %q\n", modInputs.AutoinstrumentationPath.Value, modInputs.AutoinstrumentationPreloadPath.Value)
		return nil
	}

	if !strings.EqualFold(modInputs.Backup.Value, "false") {
		if err = backupFile(modInputs.AutoinstrumentationPreloadPath.Value); err != nil {
			return err
		}
	}
	preloadFile, err := os.OpenFile(modInputs.AutoinstrumentationPreloadPath.Value, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("error opening %q: %w", modInputs.AutoinstrumentationPreloadPath.Value, err)
	}
	defer preloadFile.Close()
	if _, err = preloadFile.WriteString(modInputs.AutoinstrumentationPath.Value + "\n"); err != nil {
		return fmt.Errorf("error writing to %q: %w", modInputs.AutoinstrumentationPreloadPath.Value, err)
	}
	log.Printf("Successfully autoinstrumented preload at %q with %q\n", modInputs.AutoinstrumentationPreloadPath.Value, modInputs.AutoinstrumentationPath.Value)
	return nil
}

func RemovePreloadInstrumentation(modInputs *SplunkTAOtelLinuxAutoinstrumentationModularInputs) error {
	found, err := stringContainedInFile(modInputs.Remove.Value, modInputs.AutoinstrumentationPreloadPath.Value)
	if err != nil {
		return err
	}

	if found {
		log.Printf("Autoinstrumentation preload path %q does not exist or does not contain configured autoinstrumentation %q", modInputs.AutoinstrumentationPreloadPath.Value, modInputs.AutoinstrumentationPath.Value)
		return nil
	}

	content, err := os.ReadFile(modInputs.AutoinstrumentationPreloadPath.Value)
	if err != nil {
		return err
	}
	if !strings.EqualFold(modInputs.Backup.Value, "false") {
		if err = backupFile(modInputs.AutoinstrumentationPreloadPath.Value); err != nil {
			return err
		}
	}
	newContent := strings.ReplaceAll(string(content), modInputs.AutoinstrumentationPreloadPath.Value, "")
	return os.WriteFile(modInputs.AutoinstrumentationPreloadPath.Value, []byte(newContent), 0o644) // #nosec G306
}

func backupFile(currPath string) error {
	dir := filepath.Dir(currPath)
	origFilename := filepath.Base(currPath)
	newPathName := filepath.Join(dir, fmt.Sprintf("%s.%v", origFilename, time.Now().UnixNano()))
	if _, err := fileutils.CopyFile(currPath, newPathName); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// stringContainedInFile If the file does not exist, vacuously returns (false, nil)
func stringContainedInFile(search, filepath string) (bool, error) {
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
