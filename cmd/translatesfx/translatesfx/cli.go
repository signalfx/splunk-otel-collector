// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package translatesfx

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

// CLI is the entry point for the translatecfg command.
func CLI(args []string) {
	translateConfig(paths(args))
}

// paths returns two arguments from the command line args:
// cfgFname is the path to the Smart Agent config that is being translated
// wd is the working directory used to evaluate file paths found in the Smart Agent config
func paths(args []string) (cfgFname string, wd string) {
	switch len(args) {
	case 2:
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("error getting working directory: %v", err)
		}
		return args[1], wd
	case 3:
		return args[1], args[2]
	default:
		fmt.Println("usage: translatesacfg <path/to/smart/agent/config.yaml>")
	}
	return
}

// translateConfig takes a Smart Agent config file path and a working directory,
// then prints a translated Otel configuration to stdout.
func translateConfig(fname, wd string) {
	bytes, err := os.ReadFile(fname)
	if err != nil {
		log.Fatalf("error reading file %q: %v", fname, err)
	}
	var orig interface{}
	err = yaml.UnmarshalStrict(bytes, &orig)
	if err != nil {
		log.Fatalf("error unmarshaling file %q: %v", fname, err)
	}
	saExpanded, _ := expandSA(orig, wd)
	saInfo := saExpandedToCfgInfo(saExpanded.(map[interface{}]interface{}))
	oc := saInfoToOtelConfig(saInfo)
	bytes, err = yaml.Marshal(oc)
	if err != nil {
		log.Fatalf("error marshalling config: %v", err)
	}
	fmt.Println(string(bytes))
}
