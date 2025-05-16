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

//go:build !windows

package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/splunk/splunk-technical-addon/internal/addonruntime"
	"github.com/splunk/splunk-technical-addon/internal/modularinput"
)

func Example(mip *modularinput.ModinputProcessor) {
	splunkHome, _ := addonruntime.GetSplunkHome()
	taPlatformHome, _ := addonruntime.GetTaPlatformDir()
	taHome, _ := addonruntime.GetTaHome()

	modInputs := GetSampleAddonModularInputs(mip)
	output, err := json.Marshal(ExampleOutput{
		Flags:                      mip.GetFlags(),
		EnvVars:                    mip.GetEnvVars(),
		SplunkHome:                 splunkHome,
		TaHome:                     taHome,
		PlatformHome:               taPlatformHome,
		EverythingSet:              modInputs.EverythingSet.Value,
		UnaryFlagWithEverythingSet: modInputs.UnaryFlagWithEverythingSet.Value,
		MinimalSetRequired:         modInputs.MinimalSetRequired.Value,
		MinimalSet:                 modInputs.MinimalSet.Value,
		Platform:                   "linux",
	})
	if err != nil {
		log.Fatal("error marshalling json: ", err)
	} else {
		log.Printf("Sample output:%s\n", output)
	}
	fmt.Printf("Sample output:%s\n", output)
}
