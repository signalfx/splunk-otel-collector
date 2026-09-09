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

// ParseName splits a Splunk stanza name without applying URL semantics.
//
// Splunk uses the form kind://target for many inputs, but the target can be a
// path, an event log channel, or a network-specific value. Unprefixed names
// are returned with an empty Kind and are used by scripted inputs.
func ParseName(raw string) (Name, error) {
	if kind, target, ok := strings.Cut(raw, "://"); ok {
		if kind == "" {
			return Name{}, errEmptyKind
		}
		return Name{Kind: kind, Target: target}, nil
	}

	for _, kind := range []string{"tcp", "udp"} {
		if len(raw) > len(kind) && raw[len(kind)] == ':' && strings.EqualFold(raw[:len(kind)], kind) {
			return Name{Kind: raw[:len(kind)], Target: raw[len(kind)+1:]}, nil
		}
	}

	return Name{Target: raw}, nil
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
