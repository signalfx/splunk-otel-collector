package modularinput

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// XMLInput represents the XML input structure from Splunk
type XMLInput struct {
	XMLName       xml.Name      `xml:"input"`
	ServerHost    string        `xml:"server_host,omitempty"`
	ServerURI     string        `xml:"server_uri,omitempty"`
	SessionKey    string        `xml:"session_key,omitempty"`
	CheckpointDir string        `xml:"checkpoint_dir,omitempty"`
	Configuration Configuration `xml:"configuration"`
}

// Configuration represents the configuration block in the XML
type Configuration struct {
	XMLName xml.Name `xml:"configuration"`
	Stanzas []Stanza `xml:"stanza"`
}

// Stanza represents a stanza within the configuration
type Stanza struct {
	XMLName xml.Name `xml:"stanza"`
	Name    string   `xml:"name,attr"`
	App     string   `xml:"app,attr"`
	Params  []Param  `xml:"param"`
}

// Param represents a parameter in a stanza
type Param struct {
	XMLName xml.Name `xml:"param"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:",chardata"`
}

func UnmarshalInputXML(data []byte) (*XMLInput, error) {
	var input XMLInput
	err := xml.Unmarshal(data, &input)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling 'input.conf' XML: %w", err)
	}

	return &input, nil
}

func ReadXML(reader io.Reader) (*XMLInput, error) {
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
