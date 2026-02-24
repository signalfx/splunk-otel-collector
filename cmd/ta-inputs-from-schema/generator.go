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
	"fmt"
	"strings"
)

// generateInputsConf generates the inputs.conf file content
func generateInputsConf(scheme *Scheme, globalSettings, inputName string) string {
	var sb strings.Builder

	// Write header with input name
	sb.WriteString(fmt.Sprintf("[%s://%s]\n\n", inputName, inputName))

	// Write global settings
	sb.WriteString(globalSettings)
	if !strings.HasSuffix(globalSettings, "\n") {
		sb.WriteString("\n")
	}

	// Write TA specific settings header
	sb.WriteString("\n# TA specific settings\n")

	// Write each argument with its default value
	for _, arg := range scheme.Endpoint.Args {
		sb.WriteString(arg.Name + " = " + arg.DefaultValue + "\n")
	}

	return sb.String()
}

// generateInputsConfSpec generates the inputs.conf.spec file content
func generateInputsConfSpec(scheme *Scheme, inputName string) string {
	var sb strings.Builder

	// Write header with input name
	sb.WriteString(fmt.Sprintf("[%s://<name>]\n\n", inputName))

	// Write each argument specification
	for _, arg := range scheme.Endpoint.Args {
		// Argument name and value placeholder
		sb.WriteString(arg.Name + " = <value>\n")

		// Description (normalize whitespace)
		desc := normalizeDescription(arg.Description)

		sb.WriteString("* " + desc + "\n")

		// Default value
		sb.WriteString("* Default = " + arg.DefaultValue + "\n")

		sb.WriteString("\n")
	}

	return sb.String()
}

// normalizeDescription normalizes the description by removing extra whitespace
func normalizeDescription(desc string) string {
	// Replace multiple whitespace (including newlines) with single space
	return strings.Join(strings.Fields(desc), " ")
}
