// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkoutputsexporter

// Config holds the configuration for the splunk_outputs exporter.
type Config struct {
	// BaseDir is the Splunk installation root. outputs.conf is discovered
	// across etc/system/default, etc/apps/*, and etc/system/local using
	// standard Splunk precedence. Overrides $SPLUNK_HOME when set.
	BaseDir string `mapstructure:"base_dir"`
}
