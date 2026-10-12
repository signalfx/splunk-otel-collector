// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package stanza contains helpers for Splunk stanza names.
package stanza

import (
	"errors"
	"strings"
)

var errEmptyKind = errors.New("empty stanza kind")

// Name is the type and target encoded in a Splunk input stanza name.
// Target is intentionally kept opaque because its syntax depends on Kind.
type Name struct {
	Kind   string
	Target string
}

// ParseName splits a Splunk stanza name into its input kind and target,
// without applying URL semantics.
//
// splunkd has no central scheme registry: each input processor claims its own
// stanzas out of inputs.conf by string prefix, and those prefixes carry a
// single colon. So the split is at the first colon, and a "//" immediately
// after it belongs to the separator rather than to the target. That makes
// [monitor:/var/log] and [monitor:///var/log] the same input, and it is why
// [monitor://var/log] names a relative path.
//
// Target is kept opaque because its syntax depends on Kind: a path, an event
// log channel, or a network-specific value. Unprefixed names are returned with
// an empty Kind; inputs.conf.spec defines no unprefixed form that declares an
// input, so callers treat them as unsupported.
func ParseName(raw string) (Name, error) {
	kind, target, ok := strings.Cut(raw, ":")
	if !ok {
		return Name{Target: raw}, nil
	}
	if kind == "" {
		return Name{}, errEmptyKind
	}
	return Name{Kind: kind, Target: strings.TrimPrefix(target, "//")}, nil
}

// ParseOutputName splits a Splunk output stanza name into a kind and target.
//
// Outputs commonly use both kind:target group names, such as tcpout:primary,
// and URL-style names, such as tcpout-server://host:9997. Stanzas without a
// target use the full stanza name as the kind.
func ParseOutputName(raw string) (Name, error) {
	if raw == "" {
		return Name{}, errEmptyKind
	}
	if kind, target, ok := strings.Cut(raw, "://"); ok {
		if kind == "" {
			return Name{}, errEmptyKind
		}
		return Name{Kind: kind, Target: target}, nil
	}
	if kind, target, ok := strings.Cut(raw, ":"); ok {
		if kind == "" {
			return Name{}, errEmptyKind
		}
		return Name{Kind: kind, Target: target}, nil
	}
	return Name{Kind: raw}, nil
}

// ListenAddress converts Splunk's port-only network stanza form to the
// host:port form expected by the OpenTelemetry network receivers.
func ListenAddress(target string) string {
	if target != "" && !strings.Contains(target, ":") {
		return ":" + target
	}
	return target
}
