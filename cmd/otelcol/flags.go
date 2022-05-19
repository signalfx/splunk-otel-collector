// Copyright  Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"io"
	"strings"

	"go.opentelemetry.io/collector/service/featuregate"
)

type flags struct {
	// Command-line flags that are used by Splunk's distribution of the collector
	configFlags           *stringArrayValue
	setFlags              *stringArrayValue
	gatesList             featuregate.FlagValue
	helpFlag              bool
	noConvertConfigFlag   bool
	versionFlag           bool
	memBallastSizeMibFlag int
}

var defaultUndeclaredFlag = -1

func (f *flags) getConfigFlags() []string {
	return f.configFlags.values
}

func (f *flags) getSetFlags() []string {
	return f.setFlags.values
}

// required to support config and set flags
// taken from https://github.com/open-telemetry/opentelemetry-collector/blob/48a2e01652fa679c89259866210473fc0d42ca95/service/flags.go#L39
type stringArrayValue struct {
	values []string
}

func (s *stringArrayValue) Set(val string) error {
	s.values = append(s.values, val)
	return nil
}

func (s *stringArrayValue) String() string {
	return "[" + strings.Join(s.values, ",") + "]"
}

func (s *stringArrayValue) contains(input string) bool {
	for _, val := range s.values {
		if val == input {
			return true
		}
	}

	return false
}

func parseFlags(args []string) (flags, error) {
	flagSet := new(flag.FlagSet)
	out := flags{
		configFlags: new(stringArrayValue),
		setFlags:    new(stringArrayValue),
		gatesList:   featuregate.FlagValue{},
	}

	// This is an internal flag parser, it shouldn't give any output to user.
	flagSet.SetOutput(io.Discard)

	// Need to account for full flag names and abbreviations
	flagSet.BoolVar(&out.helpFlag, "h", false, "")
	flagSet.BoolVar(&out.helpFlag, "help", false, "")
	flagSet.BoolVar(&out.noConvertConfigFlag, "no-convert-config", false, "")
	flagSet.BoolVar(&out.versionFlag, "v", false, "")
	flagSet.BoolVar(&out.versionFlag, "version", false, "")

	// This is a deprecated option, but it is still used when set
	flagSet.IntVar(&out.memBallastSizeMibFlag, "mem-ballast-size-mib", defaultUndeclaredFlag, "")

	flagSet.Var(out.configFlags, "config", "")
	flagSet.Var(out.setFlags, "set", "")
	flagSet.Var(&out.gatesList, "feature-gates", "")

	err := flagSet.Parse(args)

	return out, err
}

func removeFlag(flags *[]string, flag string) {
	var out []string
	for _, s := range *flags {
		if s != flag {
			out = append(out, s)
		}
	}
	*flags = out
}
