// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package conf

import "sort"

// ResourceAttrs are the four Splunk resource attributes any input stanza can set.
type ResourceAttrs struct {
	Index      string
	Source     string
	Sourcetype string
	Host       string
}

// InputParams builds the stanza params for an input stanza: the resource
// attributes that are set, then every extra param sorted by name. Consumers look
// params up by name via Params.Get, so the order carries no meaning; it is fixed
// so that building the same config twice yields identical params.
func InputParams(attrs ResourceAttrs, extra map[string]string) Params {
	params := Params{}
	for _, kv := range []Param{
		{Name: "index", Value: attrs.Index},
		{Name: "source", Value: attrs.Source},
		{Name: "sourcetype", Value: attrs.Sourcetype},
		{Name: "host", Value: attrs.Host},
	} {
		if kv.Value != "" {
			params = append(params, kv)
		}
	}

	names := make([]string, 0, len(extra))
	for name := range extra {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		params = append(params, Param{Name: name, Value: extra[name]})
	}
	return params
}
