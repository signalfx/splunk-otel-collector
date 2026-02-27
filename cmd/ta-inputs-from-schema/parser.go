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
	"encoding/xml"
	"os"
)

// Scheme represents the XML scheme structure
type Scheme struct {
	XMLName     xml.Name `xml:"scheme"`
	Title       string   `xml:"title"`
	Description string   `xml:"description"`
	Endpoint    Endpoint `xml:"endpoint"`
}

// Endpoint represents the endpoint section
type Endpoint struct {
	Args []Arg `xml:"args>arg"`
}

// Arg represents an argument
type Arg struct {
	Name           string `xml:"name,attr"`
	DefaultValue   string `xml:"defaultValue,attr"`
	Title          string `xml:"title"`
	Description    string `xml:"description"`
	DataType       string `xml:"data_type"`
	RequiredOnEdit string `xml:"required_on_edit"`
}

// parseSchemeXML parses the XML scheme file
func parseSchemeXML(filename string) (*Scheme, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var scheme Scheme
	if err := xml.Unmarshal(data, &scheme); err != nil {
		return nil, err
	}

	return &scheme, nil
}

// IsRequired returns true if the argument is required
func (a *Arg) IsRequired() bool {
	return a.RequiredOnEdit == "true" && a.DefaultValue == ""
}
