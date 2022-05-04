package main

import (
	"flag"
	"io"
	"strings"

	"go.opentelemetry.io/collector/service/featuregate"
)

var (
	// Command-line flags that are used by Splunk's distribution of the collector
	helpFlag            	bool
	noConvertConfigFlag		bool
	versionFlag         	bool
	defaultUndeclaredFlag = -1
	memBallastSizeMibFlag int
	configFlags           = new(stringArrayValue)
	setFlags              = new(stringArrayValue)
	gatesList  = featuregate.FlagValue{}
)

// required to support config and set flags
// taken from https://github.com/open-telemetry/opentelemetry-collector/blob/48a2e01652fa679c89259866210473fc0d42ca95/service/flags.go#L39
type stringArrayValue struct {
	values 	[]string
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

func flags() *flag.FlagSet {
	flagSet := new(flag.FlagSet)
	// This is an internal flag parser, it shouldn't give any output to user.
	flagSet.SetOutput(io.Discard)

	// Need to account for full flag names and abbreviations
	flagSet.BoolVar(&helpFlag, "h", false, "")
	flagSet.BoolVar(&helpFlag, "help", false, "")
	flagSet.BoolVar(&noConvertConfigFlag, "no-convert-config", false, "")
	flagSet.BoolVar(&versionFlag, "v", false, "")
	flagSet.BoolVar(&versionFlag, "version", false, "")

	// This is a deprecated option, but it is still used when set
	flagSet.IntVar(&memBallastSizeMibFlag, "mem-ballast-size-mib", defaultUndeclaredFlag, "")

	flagSet.Var(configFlags, "config", "")
	flagSet.Var(setFlags, "set", "")
	flagSet.Var(gatesList, "feature-gates", "")

	return flagSet
}

func getConfigFlags() []string {
	return configFlags.values
}

func getSetFlags() []string {
	return setFlags.values
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
