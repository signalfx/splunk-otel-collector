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
	XMLName        xml.Name      `xml:"input"`
	ServerHost     string        `xml:"server_host"`
	ServerURI      string        `xml:"server_uri"`
	SessionKey     string        `xml:"session_key"`
	CheckpointDir  string        `xml:"checkpoint_dir"`
	Configuration  Configuration `xml:"configuration"`
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
