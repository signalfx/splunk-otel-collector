// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package stanza

import "strings"

// builtinKinds are the input schemes splunkd compiles in, the `builtin_schemes`
// array in SpecFiles.cpp. The set matters beyond validation: splunkd assumes any
// stanza whose scheme is not in it comes from a modular input
// (PropertyPagesInternal::is_modinput), and a modular input's defaults come from
// its own [<scheme>] stanza rather than [default].
var builtinKinds = map[string]bool{
	"batch":         true,
	"fifo":          true,
	"filter":        true,
	"fschange":      true,
	"http":          true,
	"monitor":       true,
	"script":        true,
	"splunktcp-ssl": true,
	"splunktcp":     true,
	"ssl":           true,
	"tcp-ssl":       true,
	"tcp":           true,
	"udp":           true,
}

// IsBuiltinKind reports whether kind is one of splunkd's built-in input
// schemes. Matching is case-insensitive, as splunkd's lookup is (it bsearches
// with strcasecmp), so [MONITOR://x] is the built-in monitor input.
func IsBuiltinKind(kind string) bool {
	return builtinKinds[strings.ToLower(kind)]
}

// IsModularInput reports whether splunkd would read this stanza as belonging to
// a modular input, which decides where its defaults come from.
//
// splunkd looks for the first ":" rather than "://" because several built-in
// stanzas separate with a bare colon ([tcp:9997], [blacklist:<path>]), and those
// must not be taken for modular inputs. A name with a colon but no "://" is
// therefore never one, and a name with no colon at all is the scheme on its own,
// which is how a scheme's defaults stanza is written.
func IsModularInput(name string) bool {
	kind, rest, hasColon := strings.Cut(name, ":")
	if hasColon && !strings.HasPrefix(rest, "//") {
		return false
	}
	if !hasColon {
		kind = name
	}
	return !IsBuiltinKind(kind)
}
