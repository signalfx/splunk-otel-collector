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

//go:build integration

package tests

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	docker "github.com/docker/docker/client"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestTelegrafProcstatReceiverProvidesAllMetrics(t *testing.T) {
	expectedMetrics := "all.yaml"
	// telegraf/procstat is missing cpu metrics on arm64 as an apparently unsupported platform.
	// This arch check should be made available to a helper as similar differences are discovered
	image := os.Getenv("SPLUNK_OTEL_COLLECTOR_IMAGE")
	if strings.TrimSpace(image) != "" {
		client, err := docker.NewClientWithOpts(docker.FromEnv)
		require.NoError(t, err)
		client.NegotiateAPIVersion(context.Background())
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		inspect, _, err := client.ImageInspectWithRaw(ctx, image)
		require.NoError(t, err)
		if inspect.Architecture == "arm64" {
			expectedMetrics = "arm64.yaml"
		}
	}
	testutils.AssertAllMetricsReceived(t, expectedMetrics, "all_metrics_config.yaml", nil, nil)
}
