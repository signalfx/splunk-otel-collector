// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package conf

// Configuration holds a stanza configuration.
type Configuration struct {
	Stanza Stanza `xml:"stanza"`
}

// Params is a slice of Param.
type Params []Param

// Get returns the parameter with the given name, or nil if not found.
func (p Params) Get(name string) *Param {
	for _, param := range p {
		if param.Name == name {
			return &param
		}
	}
	return nil
}

// Stanza represents a configuration stanza.
type Stanza struct {
	Name   string `xml:"name,attr"`
	App    string `xml:"app,attr"`
	Params Params `xml:"param"`
}

// Param represents a parameter within a stanza.
type Param struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",innerxml"`
}
