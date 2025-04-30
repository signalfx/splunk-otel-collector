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
)

func SetupAddonLogger(logFilePath string) (closer func(), err error) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.LUTC | log.Lshortfile)

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
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
	if !ok {
		ex, err := os.Executable()
		if err != nil {
			// Don't attempt to further figure out path
			return "", err
		}
		// $SPLUNK_HOME/etc/apps/Sample_Addon/${platform}_x86_64/bin/executable
		splunkHome = filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(ex)))))
	}

	return filepath.Join(splunkHome, "var", "log", "splunk", schemaName+".log"), nil
}
