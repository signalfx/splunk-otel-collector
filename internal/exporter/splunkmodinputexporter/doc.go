// Copyright Splunk Inc. 2025
// SPDX-License-Identifier: Apache-2.0

//go:build mdatagen

//go:generate mdatagen metadata.yaml

// Package splunkmodinputexporter exports telemetry as newline-delimited Splunk HEC JSON
// events to standard output.
package splunkmodinputexporter // import "github.com/signalfx/splunk-otel-collector/internal/exporter/splunkmodinputexporter"
