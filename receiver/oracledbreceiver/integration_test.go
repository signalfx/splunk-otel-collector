// Copyright Splunk, Inc.
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

//go:build integration
// +build integration

package oracledbreceiver

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
)

// This test ensures the collector can connect to an Oracle DB, and properly get metrics. It's not intended to
// test the receiver itself.
func TestOracleDBIntegration(t *testing.T) {
	externalPort := "51521"
	internalPort := "1521"

	// The Oracle DB container takes close to 10 minutes on a local machine to do the default setup, so the best way to
	// account for startup time is to wait for the container to be healthy before continuing test.
	waitStrategy := wait.NewHealthStrategy().WithStartupTimeout(15 * time.Minute)
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    filepath.Join("testdata", "integration"),
			Dockerfile: "Dockerfile.oracledb",
		},
		ExposedPorts: []string{externalPort + ":" + internalPort},
		WaitingFor:   waitStrategy,
	}
	ctx := context.Background()

	container, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		},
	)
	require.NotNil(t, container)
	require.NoError(t, err)

	factory := NewFactory()
	config := factory.CreateDefaultConfig().(*Config)
	config.DataSource = "oracle://otel:password@localhost:51521/XE"
	consumer := &consumertest.MetricsSink{}
	receiver, err := factory.CreateMetricsReceiver(
		ctx,
		componenttest.NewNopReceiverCreateSettings(),
		config,
		consumer,
	)
	require.NoError(t, err)
	err = receiver.Start(ctx, componenttest.NewNopHost())
	require.NoError(t, err)
	require.Eventuallyf(
		t,
		func() bool {
			return consumer.DataPointCount() > 0
		},
		15*time.Minute,
		1*time.Second,
		"failed to receive more than 0 metrics",
	)
}
