// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package conf

import "strings"

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

// Bool reports the boolean splunkd would read from the named param, leaving def
// in place for a missing param or an unrecognized value.
func (p Params) Bool(name string, def bool) bool {
	param := p.Get(name)
	if param == nil {
		return def
	}
	switch strings.ToLower(param.Value) {
	case "f", "false", "no", "0":
		return false
	case "t", "true", "yes", "1":
		return true
	}
	return def
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
