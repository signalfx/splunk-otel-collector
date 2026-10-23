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

package cli

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// Exit codes follow the LSB init-script and systemctl conventions so scripts
// can branch on them. Every command family should use these rather than
// inventing its own.
const (
	ExitOK          = 0
	ExitFailure     = 1
	ExitUsage       = 2
	ExitNotRunning  = 3 // LSB "program is not running"; matches systemctl status.
	ExitUnsupported = 4
)

// IO carries the streams a command family writes to.
type IO struct {
	Out io.Writer
	Err io.Writer
}

// VerbInfo describes one bare command name a Family recognizes at argv[0],
// along with a one-line description for --help/help output.
type VerbInfo struct {
	Name string
	Help string
}

// Family is a group of related top-level otelcollauncher subcommands (e.g.
// lifecycle management: start/stop/status/restart).
type Family interface {
	Verbs() []VerbInfo
	Dispatch(verb string, args []string, streams IO) int
}

var helpVerbs = map[string]bool{"help": true, "-h": true, "--help": true}

// Dispatch matches args[0] against the verbs of every registered family, in
// order, and runs the first match. If two families register the same verb, the
// first one registered wins.
func Dispatch(families []Family, args []string, streams IO) (code int, ok bool) {
	if len(args) == 0 {
		return 0, false
	}
	if helpVerbs[args[0]] {
		fmt.Fprint(streams.Out, HelpText(families))
		return ExitOK, true
	}
	if strings.HasPrefix(args[0], "-") {
		return 0, false
	}
	for _, f := range families {
		for _, v := range f.Verbs() {
			if v.Name == args[0] {
				return f.Dispatch(v.Name, args[1:], streams), true
			}
		}
	}
	return 0, false
}

// HelpText renders the auto-generated command list from every registered
// family's Verbs().
func HelpText(families []Family) string {
	var b strings.Builder
	b.WriteString("Usage: otelcollauncher [command] [args...]\n\n")

	var verbs []VerbInfo
	for _, f := range families {
		verbs = append(verbs, f.Verbs()...)
	}

	if len(verbs) > 0 {
		b.WriteString("Commands:\n")
		tw := tabwriter.NewWriter(&b, 0, 0, 3, ' ', 0)
		for _, v := range verbs {
			fmt.Fprintf(tw, "  %s\t%s\n", v.Name, v.Help)
		}
		_ = tw.Flush()
		b.WriteString("\n")
	}

	return b.String()
}
