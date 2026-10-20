// Copyright Splunk Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package supportbundle

import (
	"context"
	"flag"
	"fmt"

	"github.com/signalfx/splunk-otel-collector/cmd/otelcollauncher/cli"
	"github.com/signalfx/splunk-otel-collector/internal/supportbundle"
)

const verbName = "support-bundle"

const usageText = `Collect troubleshooting data into an archive for the Splunk OpenTelemetry Collector.

Flags:
  -c, --config-dir string  Collector configuration directory to include (defaults to the platform collector configuration directory).
  -t, --tmpdir string      Temporary working directory for bundle contents; it must not exist (defaults to a timestamped directory under the platform temp root).

Examples:
  otelcollauncher support-bundle
  otelcollauncher support-bundle --config-dir /etc/otel/collector
  otelcollauncher support-bundle --tmpdir /tmp/support-bundle
`

type Family struct{}

// New returns the support-bundle command family.
func New() *Family {
	return &Family{}
}

func (Family) Verbs() []cli.VerbInfo {
	return []cli.VerbInfo{
		{
			Name:  verbName,
			Help:  "collect troubleshooting data into a support bundle",
			Usage: usageText,
		},
	}
}

func (Family) Dispatch(verb string, args []string, streams cli.IO) int {
	if verb != verbName {
		return cli.ExitUsage
	}
	return run(args, streams)
}

type options struct {
	configDir string
	tempDir   string
}

func parseArgs(args []string, streams cli.IO) (options, int, bool) {
	fs := flag.NewFlagSet(verbName, flag.ContinueOnError)
	fs.SetOutput(streams.Err)
	fs.Usage = func() { fmt.Fprint(streams.Err, usageText) }

	var opts options
	fs.StringVar(&opts.configDir, "config-dir", "", "collector configuration directory")
	fs.StringVar(&opts.configDir, "c", "", "collector configuration directory (shorthand for --config-dir)")
	fs.StringVar(&opts.tempDir, "tmpdir", "", "temporary working directory for bundle contents")
	fs.StringVar(&opts.tempDir, "t", "", "temporary working directory for bundle contents (shorthand for --tmpdir)")

	if err := fs.Parse(args); err != nil {
		return options{}, cli.ExitUsage, false
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(streams.Err, "%s: unexpected argument %q\n\n", verbName, fs.Arg(0))
		fmt.Fprint(streams.Err, usageText)
		return options{}, cli.ExitUsage, false
	}
	return opts, cli.ExitOK, true
}

func run(args []string, streams cli.IO) int {
	opts, code, ok := parseArgs(args, streams)
	if !ok {
		return code
	}

	if _, err := supportbundle.Run(context.Background(), supportbundle.Options{
		ConfigDir: opts.configDir,
		TempDir:   opts.tempDir,
		Stdout:    streams.Out,
	}); err != nil {
		fmt.Fprintf(streams.Err, "support-bundle: %v\n", err)
		return cli.ExitFailure
	}
	return cli.ExitOK
}
