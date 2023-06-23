// Copyright Splunk, Inc.
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

package testutils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/signalfx/splunk-otel-collector/tests/testutils/subprocess"
)

const binaryPathSuffix = "/bin/otelcol"
const findExecutableErrorMsg = "unable to find collector executable path.  Be sure to run `make otelcol`"

var _ Collector = (*CollectorProcess)(nil)

type CollectorProcess struct {
	Env              map[string]string
	Logger           *zap.Logger
	Process          *subprocess.Subprocess
	subprocessConfig *subprocess.Config
	Path             string
	ConfigPath       string
	LogLevel         string
	Args             []string
	Fail             bool
}

func (collector CollectorProcess) WithMount(string, string) Collector {
	return &collector
}

// To be used as a builder whose Build() method provides the actual instance capable of launching the process.
func NewCollectorProcess() CollectorProcess {
	return CollectorProcess{Env: map[string]string{}}
}

// Nearest `bin/otelcol` by default
func (collector CollectorProcess) WithPath(path string) CollectorProcess {
	collector.Path = path
	return collector
}

// Required
func (collector CollectorProcess) WithConfigPath(path string) Collector {
	collector.ConfigPath = path
	return &collector
}

// []string{"--set=service.telemetry.logs.level={collector.LogLevel}", "--config", collector.ConfigPath, "--set=service.telemetry.metrics.level=none"} by default
func (collector CollectorProcess) WithArgs(args ...string) Collector {
	collector.Args = args
	return &collector
}

// empty by default
func (collector CollectorProcess) WithEnv(env map[string]string) Collector {
	for k, v := range env {
		collector.Env[k] = v
	}
	return &collector
}

// Nop logger by default
func (collector CollectorProcess) WithLogger(logger *zap.Logger) Collector {
	collector.Logger = logger
	return &collector
}

// info by default
func (collector CollectorProcess) WithLogLevel(level string) Collector {
	collector.LogLevel = level
	return &collector
}

// noop at this time
func (collector CollectorProcess) WillFail(fail bool) Collector {
	collector.Fail = fail
	return &collector
}

func (collector CollectorProcess) Build() (Collector, error) {
	if collector.Path == "" {
		collectorPath, err := findCollectorPath()
		if err != nil {
			return nil, err
		}
		collector.Path = collectorPath
	}
	if collector.Logger == nil {
		collector.Logger = zap.NewNop()
	}
	if collector.LogLevel == "" {
		collector.LogLevel = "info"
	}

	configSetByArgs := configIsSetByArgs(collector.Args)
	_, configSetByEnvVar := collector.Env["SPLUNK_CONFIG"]
	if collector.ConfigPath != "" && !configSetByArgs && !configSetByEnvVar {
		// only specify w/ args if none are used in the test
		if collector.Args == nil {
			collector.Args = []string{
				fmt.Sprintf("--set=service.telemetry.logs.level=%s", collector.LogLevel), "--config", collector.ConfigPath, "--set=service.telemetry.metrics.level=none",
			}
		} else {
			// fallback to env var
			collector.Env["SPLUNK_CONFIG"] = collector.ConfigPath
		}
	}

	collector.subprocessConfig = &subprocess.Config{
		ExecutablePath:       collector.Path,
		Args:                 collector.Args,
		EnvironmentVariables: collector.Env,
	}
	collector.Process = subprocess.NewSubprocess(collector.subprocessConfig, collector.Logger)
	return &collector, nil
}

func (collector *CollectorProcess) Start() error {
	if collector.Process == nil {
		return fmt.Errorf("cannot Start a CollectorProcess that hasn't been successfully built")
	}
	go func() {
		// drain stdout/err buffer (already logged for us)
		for range collector.Process.Stdout {
			continue
		}
	}()

	return collector.Process.Start(context.Background())
}

func (collector *CollectorProcess) Shutdown() error {
	if collector.Process == nil {
		return fmt.Errorf("cannot Shutdown a CollectorProcess that hasn't been successfully built")
	}

	return collector.Process.Shutdown(context.Background())
}

func (collector *CollectorProcess) InitialConfig(t testing.TB, port uint16) map[string]any {
	return requestConfig(t, fmt.Sprintf("http://localhost:%d/debug/configz/initial", port))
}

func (collector *CollectorProcess) EffectiveConfig(t testing.TB, port uint16) map[string]any {
	return requestConfig(t, fmt.Sprintf("http://localhost:%d/debug/configz/effective", port))
}

func requestConfig(t testing.TB, uri string) map[string]any {
	var resp *http.Response
	var err error
	for i := 0; i < 3; i++ {
		//nolint:gosec
		resp, err = http.Get(uri) // uri cannot be constant since ports are variable
		if err == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	actual := map[string]any{}
	require.NoError(t, yaml.Unmarshal(body, &actual))
	return confmap.NewFromStringMap(actual).ToStringMap()
}

// Walks up parent directories looking for bin/otelcol
func findCollectorPath() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	binaryPath := binaryPathSuffix
	var collectorPath string
	for i := 0; true; i++ {
		attemptedPath := path.Join(dir, binaryPath)
		var info os.FileInfo
		if info, err = os.Stat(attemptedPath); err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("%s: %w", findExecutableErrorMsg, err)
		}

		if info != nil {
			collectorPath = attemptedPath
			break
		}
		if attemptedPath == binaryPathSuffix {
			break // at root
		}
		dir = path.Join(dir, "..")
	}

	if collectorPath == "" {
		err = fmt.Errorf(findExecutableErrorMsg)
	}
	return collectorPath, err
}
