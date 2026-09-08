// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkinputsreceiver

type Config struct {
	// BaseDir is the Splunk installation root ($SPLUNK_HOME). TAs are
	// discovered under etc/apps/* and conf files are layered using standard
	// Splunk precedence. Falls back to the $SPLUNK_HOME environment variable
	// when not set.
	BaseDir string `mapstructure:"base_dir"`
}
