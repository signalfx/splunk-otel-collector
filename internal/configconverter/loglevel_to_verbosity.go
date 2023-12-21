// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package configconverter

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"go.opentelemetry.io/collector/config/configtelemetry"
	"go.opentelemetry.io/collector/confmap"
	"go.uber.org/zap/zapcore"
)

type LogLevelToVerbosity struct{}

func (LogLevelToVerbosity) Convert(_ context.Context, in *confmap.Conf) error {
	if in == nil {
		return fmt.Errorf("cannot MoveHecTLS on nil *confmap.Conf")
	}

	const expression = "exporters::logging(/\\w+)?::loglevel"
	re := regexp.MustCompile(expression)
	out := map[string]any{}
	unsupportedKeyFound := false
	for _, k := range in.AllKeys() {
		v := in.Get(k)
		match := re.FindStringSubmatch(k)
		if match == nil {
			out[k] = v
		} else {
			// check if verbosity is also set:
			verbosityKey := fmt.Sprintf("exporters::logging%s::verbosity", match[1])
			if in.Get(verbosityKey) == nil {
				log.Printf("Deprecated key found: %s. Translating to %s\n", k, verbosityKey)
				var l zapcore.Level
				if err := l.UnmarshalText([]byte(v.(string))); err != nil {
					log.Printf("Could not read loglevel: %v", err)
					out[k] = v
				} else {
					var verbosityLevel configtelemetry.Level
					if verbosityLevel, err = mapLevel(l); err != nil {
						log.Printf("Could not map loglevel to verbosity: %v", err)
						out[k] = v
					} else {
						out[verbosityKey] = verbosityLevel.String()
					}
				}
			} else {
				log.Printf("Deprecated key found: %s. Found new key %s", k, verbosityKey)
			}
			unsupportedKeyFound = true
		}
	}
	if unsupportedKeyFound {
		log.Println(
			"[WARNING] `exporters` -> `logging` -> `loglevel` " +
				"is deprecated and moved to verbosity. Please update your config. " +
				"https://github.com/open-telemetry/opentelemetry-collector/pull/6334",
		)
	}

	*in = *confmap.NewFromStringMap(out)
	return nil
}

func mapLevel(level zapcore.Level) (configtelemetry.Level, error) {
	switch level {
	case zapcore.DebugLevel:
		return configtelemetry.LevelDetailed, nil
	case zapcore.InfoLevel:
		return configtelemetry.LevelNormal, nil
	case zapcore.WarnLevel, zapcore.ErrorLevel,
		zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		// Anything above info is mapped to 'basic' level.
		return configtelemetry.LevelBasic, nil
	default:
		return configtelemetry.LevelNone, fmt.Errorf("log level %q is not supported", level)
	}
}
