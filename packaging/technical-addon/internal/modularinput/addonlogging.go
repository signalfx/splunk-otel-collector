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

package modularinput

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/splunk/splunk-technical-addon/internal/addonruntime"
)

func SetupAddonLogger(logFilePath string) (closer func(), err error) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.LUTC | log.Lshortfile)

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		log.Printf("Error: failed to open log file %s: %v", logFilePath, err)
		return func() {}, fmt.Errorf("could not open log file %s", logFilePath)
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)

	return func() { _ = logFile.Close() }, nil
}

func GetDefaultLogFilePath(schemaName string) (string, error) {
	splunkHome, ok := os.LookupEnv("SPLUNK_HOME")
	var err error
	if !ok {
		if splunkHome, err = addonruntime.GetSplunkHome(); err != nil {
			return "", err
		}
	}

	return filepath.Join(splunkHome, "var", "log", "splunk", schemaName+".log"), nil
}
