// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package conf

type Configuration struct {
	Stanza Stanza `xml:"stanza"`
}

type Params []Param

func (p Params) Get(name string) *Param {
	for _, param := range p {
		if param.Name == name {
			return &param
		}
	}
	return nil
}

type Stanza struct {
	Name   string `xml:"name,attr"`
	App    string `xml:"app,attr"`
	Params Params `xml:"param"`
}

type Param struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",innerxml"`
}
