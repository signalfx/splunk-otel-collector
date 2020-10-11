// Copyright 2020 Splunk, Inc.
// Copyright The OpenTelemetry Authors
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

// Wrapper to otelcol program.
package main

import (
	"log"
	"os"
	"os/exec"
)

const (
	SPLUNK_CONFIG = "/etc/otel/collector/splunk_config.yaml"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	config := SPLUNK_CONFIG
	envVars := []string{"$SPLUNK_TOKEN", "$SPLUNK_REALM", "$SPLUNK_BALLAST"}
	for _, v := range envVars {
		if len(os.ExpandEnv(v)) == 0 {
			log.Fatalf("missing environment variable %s", v)
		}
	}
	_, ok := os.LookupEnv("SPLUNK_CONFIG")
	if ok {
		config = os.ExpandEnv("$SPLUNK_CONFIG")
	}

	args := os.Args[1:]

	if len(args) == 0 {
		_, err := os.Stat(config)
		if os.IsNotExist(err) {
			log.Fatalf("unable to find configuration file (%s) ensure SPLUNK_CONFIG environment variable is set properly", config)
		}
		log.Printf("Running: ./otelcol --config %s --mem-ballast-size-mib %s", config, os.ExpandEnv("$SPLUNK_BALLAST"))
		cmd := exec.Command("./otelcol", "--config", config, "--mem-ballast-size-mib", os.ExpandEnv("$SPLUNK_BALLAST"))
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			log.Fatalln(err)
		}

	} else {
		log.Println("Running: ./otelcol", args)
		cmd := exec.Command("./otelcol", args...)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			log.Fatalln(err)
		}
	}
}
