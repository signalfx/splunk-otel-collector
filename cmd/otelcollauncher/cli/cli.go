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

// Follow the LSB init-script and systemctl Exit codes conventions
// Every command family should use these Exit codes rather than inventing its own.
const (
	ExitOK          = 0
	ExitFailure     = 1
	ExitUsage       = 2
	ExitNotRunning  = 3 // LSB "program is not running"; matches systemctl status.
	ExitUnsupported = 4
)

type IO struct {
	Out io.Writer
	Err io.Writer
}

// VerbInfo describes one bare command name a Family recognizes at argv[0].
// Help is a one-line synopsis shown in the top-level command list. Usage is
// optional, extended help shown for that verb alone (via `help <verb>` or
// `<verb> --help`/`<verb> -h`); when empty, per-verb help falls back to
// showing just Help.
type VerbInfo struct {
	Name  string
	Help  string
	Usage string
}

// Family is a group of related top-level otelcollauncher subcommands (e.g.
// lifecycle management: start/stop/status/restart).
type Family interface {
	Verbs() []VerbInfo
	Dispatch(verb string, args []string, streams IO) int
}

// helpVerb is the only verb that triggers the launcher's own help
// output. -h and --help are deliberately NOT intercepted here: before any
// command family existed, otelcollauncher forwarded every "-"-prefixed
// argument straight through to the collector or supervisor (e.g.
// `otelcollauncher --help` showed otelcol's own flags), and that passthrough
// contract must hold with no exceptions.
const helpVerb = "help"

// Dispatch matches args[0] against the verbs of every registered family, in
// order, and runs the first match. If two families register the same verb, the
// first one registered wins.
func Dispatch(families []Family, args []string, streams IO) (code int, ok bool) {
	if len(args) == 0 {
		return 0, false
	}
	if args[0] == helpVerb {
		return dispatchHelp(families, args[1:], streams), true
	}
	if strings.HasPrefix(args[0], "-") {
		return 0, false
	}
	for _, f := range families {
		for _, v := range f.Verbs() {
			if v.Name != args[0] {
				continue
			}
			rest := args[1:]
			if wantsVerbHelp(rest) {
				fmt.Fprint(streams.Out, verbHelpText(v))
				return ExitOK, true
			}
			return f.Dispatch(v.Name, rest, streams), true
		}
	}
	return 0, false
}

func wantsVerbHelp(args []string) bool {
	return len(args) > 0 && (args[0] == "-h" || args[0] == "--help")
}

func dispatchHelp(families []Family, args []string, streams IO) int {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		fmt.Fprint(streams.Out, HelpText(families))
		return ExitOK
	}
	if v, found := findVerb(families, args[0]); found {
		fmt.Fprint(streams.Out, verbHelpText(v))
		return ExitOK
	}
	fmt.Fprintf(streams.Err, "unknown command %q\n\n", args[0])
	fmt.Fprint(streams.Out, HelpText(families))
	return ExitUsage
}

func findVerb(families []Family, name string) (VerbInfo, bool) {
	for _, f := range families {
		for _, v := range f.Verbs() {
			if v.Name == name {
				return v, true
			}
		}
	}
	return VerbInfo{}, false
}

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

func verbHelpText(v VerbInfo) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Usage: otelcollauncher %s\n\n%s\n", v.Name, v.Help)
	if v.Usage != "" {
		b.WriteString("\n")
		b.WriteString(strings.TrimRight(v.Usage, "\n"))
		b.WriteString("\n")
	}
	return b.String()
}
