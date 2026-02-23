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

package modularinput

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

type Param struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

type Stanza struct {
	Name  string  `xml:"name,attr"`
	App   string  `xml:"app,attr"`
	Param []Param `xml:"param"`
}

type Configuration struct {
	Stanza []Stanza `xml:"stanza"`
}

type Input struct {
	XMLName       xml.Name      `xml:"input"`
	ServerHost    string        `xml:"server_host"`
	ServerURI     string        `xml:"server_uri"`
	SessionKey    string        `xml:"session_key"`
	CheckpointDir string        `xml:"checkpoint_dir"`
	Configuration Configuration `xml:"configuration"`
}

func UnmarshalInputXML(data []byte) (*Input, error) {
	var input Input
	err := xml.Unmarshal(data, &input)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling 'input.conf' XML: %w", err)
	}

	return &input, nil
}

func ReadXML(reader io.Reader) (*Input, error) {
	var xmlData strings.Builder
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()
		xmlData.WriteString(line)
		if strings.Contains(line, "</input>") {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading input: %w", err)
	}

	return UnmarshalInputXML([]byte(xmlData.String()))
}
