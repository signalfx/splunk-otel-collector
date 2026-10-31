// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build mdatagen

//go:generate mdatagen metadata.yaml

// Package splunkauthextension authenticates receiver requests using Splunk HEC tokens.
package splunkauthextension
