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

//go:build installation

package tests

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"testing"
	"time"

	dockerContainer "github.com/docker/docker/api/types/container"
	dockerMount "github.com/docker/docker/api/types/mount"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func getPackagePath(t testing.TB, suffix string) string {
	var paths []string
	require.NoError(
		t, filepath.Walk(
			filepath.Join("..", "..", "dist"),
			func(path string, info os.FileInfo, err error) error {
				file := filepath.Base(path)
				if strings.HasPrefix(file, "splunk-otel-collector") && strings.HasSuffix(file, fmt.Sprintf(".%s", suffix)) {
					abs, e := filepath.Abs(path)
					require.NoError(t, e)
					paths = append(paths, abs)
				}
				return nil
			}),
	)
	if len(paths) == 0 {
		t.Fatalf("no %s installer available: run `make %s-package` for this test.", suffix, suffix)
	}

	sort.Strings(paths)
	return paths[len(paths)-1]
}

func TestDefaultConfigDDiscoversPostgres(t *testing.T) {
	tc := testutils.NewTestcase(t, testutils.OTLPReceiverSinkBindToBridgeGateway)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	finfo, err := os.Stat("/var/run/docker.sock")
	require.NoError(t, err)
	fsys := finfo.Sys()
	stat, ok := fsys.(*syscall.Stat_t)
	require.True(t, ok)
	dockerGID := fmt.Sprintf("%d", stat.Gid)

	treeBytes, err := os.ReadFile(filepath.Join(".", "testdata", "tree"))
	require.NoError(t, err)
	expectedTree := string(treeBytes)

	server := testutils.NewContainer().WithContext(filepath.Join("..", "receivers", "smartagent", "postgresql", "testdata", "server")).WithEnv(
		map[string]string{"POSTGRES_DB": "test_db", "POSTGRES_USER": "postgres", "POSTGRES_PASSWORD": "postgres"},
	).WithExposedPorts("5432:5432").WithName("postgres-server").WithNetworks(
		"postgres",
	).WillWaitForPorts("5432").WillWaitForLogs("database system is ready to accept connections")

	client := testutils.NewContainer().WithContext(filepath.Join("..", "receivers", "smartagent", "postgresql", "testdata", "client")).WithEnv(
		map[string]string{"POSTGRES_SERVER": "postgres-server"},
	).WithName("postgres-client").WithNetworks("postgres").WillWaitForLogs("Beginning psql requests")

	_, stop := tc.Containers(server, client)
	defer stop()

	waitForSystemd := wait.NewExecStrategy([]string{"systemctl", "status", "1"})
	waitForSystemd.PollInterval = time.Second
	waitForSystemd.WithResponseMatcher(
		func(body io.Reader) bool {
			b, err := io.ReadAll(body)
			require.NoError(t, err)
			return strings.Contains(string(b), "Active: active (running)")
		})

	for _, packageType := range []string{"deb", "rpm"} {
		t.Run(packageType, func(t *testing.T) {
			defer tc.OTLPReceiverSink.Reset()
			packagePath := getPackagePath(t, packageType)
			cc, shutdown := tc.SplunkOtelCollectorContainer(
				"", func(c testutils.Collector) testutils.Collector {
					cc := c.(*testutils.CollectorContainer)
					cc.Container.Privileged = true
					cc.Container = cc.Container.WithContext(
						filepath.Join(".", "testdata", "systemd"),
					).WithDockerfile(
						fmt.Sprintf("Dockerfile.%s", packageType),
					).WithHostConfigModifier(func(config *dockerContainer.HostConfig) {
						config.CgroupnsMode = dockerContainer.CgroupnsModeHost
						config.CapAdd = append(config.CapAdd, "NET_RAW")
						config.GroupAdd = []string{dockerGID}
						config.Mounts = []dockerMount.Mount{
							{Source: "/sys/fs/cgroup", Target: "/sys/fs/cgroup", Type: dockerMount.TypeBind},
							{Source: packagePath, Target: fmt.Sprintf("/opt/otel/splunk-otel-collector.%s", packageType), Type: dockerMount.TypeBind},
							{Source: "/var/run/docker.sock", Target: "/opt/docker/docker.sock", ReadOnly: true, Type: dockerMount.TypeBind},
						}
					}).WithBuildArgs(map[string]*string{"DOCKER_GID": &dockerGID}).WithNetworks("postgres")
					cc.Container.WillWaitForLogs()
					cc.Container.WaitingFor = append(cc.Container.WaitingFor, waitForSystemd)
					return cc.WithArgs("")
				},
			)

			var installCmd []string
			switch packageType {
			case "deb":
				installCmd = []string{"apt-get", "install", "-f", "/opt/otel/splunk-otel-collector.deb", "-y"}
			case "rpm":
				installCmd = []string{"dnf", "install", "/opt/otel/splunk-otel-collector.rpm", "-y"}
			}

			defer shutdown()

			for _, exec := range []struct {
				wait time.Duration
				cmd  []string
			}{
				{wait: 5 * time.Minute, cmd: installCmd},
				{wait: 10 * time.Second, cmd: []string{"usermod", "-aG", "docker", "splunk-otel-collector"}},
				{wait: 20 * time.Second, cmd: []string{"systemctl", "import-environment", "OTLP_ENDPOINT"}},
				{wait: 20 * time.Second, cmd: []string{"systemctl", "daemon-reload"}},
				{wait: 10 * time.Second, cmd: []string{"systemctl", "start", "splunk-otel-collector"}},
			} {
				rc, stdout, stderr := cc.Container.AssertExec(t, exec.wait, exec.cmd...)
				require.Zero(t, rc, fmt.Sprintf("%s failed. stdout: %q, stderr: %q", strings.Join(exec.cmd, " "), stdout, stderr))
			}

			// --discovery default time
			time.Sleep(10 * time.Second)

			expectedResourceMetrics := tc.ResourceMetrics("postgres-and-internal.yaml")
			require.NoError(t, tc.OTLPReceiverSink.AssertAllMetricsReceived(t, *expectedResourceMetrics, 30*time.Second))

			expectedInitial := map[string]any{
				"file": map[string]any{},
				"splunk.configd": map[string]any{
					"exporters": map[string]any{
						"otlp": map[string]any{
							"endpoint": "${env:OTLP_ENDPOINT}",
							"tls": map[string]any{
								"insecure": true,
							},
						},
					},
					"extensions": map[string]any{},
					"processors": map[string]any{},
					"receivers": map[string]any{
						"prometheus/internal": map[string]any{
							"config": map[string]any{
								"scrape_configs": []any{
									map[string]any{
										"job_name": "otel-collector",
										"metric_relabel_configs": []any{
											map[string]any{
												"action":        "drop",
												"regex":         ".*grpc_io.*",
												"source_labels": []any{"__name__"},
											},
										},
										"scrape_interval": "10s",
										"static_configs": []any{
											map[string]any{
												"targets": []any{"0.0.0.0:8888"},
											},
										},
									},
								},
							},
						},
					},
					"service": map[string]any{
						"pipelines": map[string]any{
							"metrics": map[string]any{
								"exporters": []any{"otlp"},
								"receivers": []any{"prometheus/internal"},
							},
						},
					},
				},
				"splunk.discovery": map[string]any{
					"extensions": map[string]any{
						"docker_observer": map[string]any{
							"endpoint": "unix:///opt/docker/docker.sock",
						},
					},
					"receivers": map[string]any{
						"receiver_creator/discovery": map[string]any{
							"receivers": map[string]any{
								"smartagent/postgresql": map[string]any{
									"config": map[string]any{
										"connectionString": "sslmode=disable user={{.username}} password={{.password}}",
										"masterDBName":     "test_db",
										"params": map[string]any{
											"password": "test_password",
											"username": "test_user",
										},
										"type": "postgresql",
									},
									"resource_attributes": map[string]any{},
									"rule":                "type == \"container\" and any([name, image, command], {# matches \"(?i)postgres\"}) and not (command matches \"splunk.discovery\")",
								},
							},
							"watch_observers": []any{"docker_observer"}}},
					"service": map[string]any{
						"extensions/splunk.discovery": []any{"docker_observer"},
						"receivers/splunk.discovery":  []any{"receiver_creator/discovery"},
					},
				},
			}
			require.Equal(t, expectedInitial, cc.InitialConfig(t, 55554))

			expectedEffective := map[string]any{
				"exporters": map[string]any{
					"otlp": map[string]any{
						"endpoint": tc.OTLPEndpoint,
						"tls": map[string]any{
							"insecure": true,
						},
					},
				},
				"extensions": map[string]any{
					"docker_observer": map[string]any{
						"endpoint": "unix:///opt/docker/docker.sock",
					},
				},
				"processors": map[string]any{},
				"receivers": map[string]any{
					"prometheus/internal": map[string]any{
						"config": map[string]any{
							"scrape_configs": []any{
								map[string]any{
									"job_name": "otel-collector",
									"metric_relabel_configs": []any{
										map[string]any{
											"action":        "drop",
											"regex":         ".*grpc_io.*",
											"source_labels": []any{"__name__"},
										},
									},
									"scrape_interval": "10s",
									"static_configs": []any{
										map[string]any{
											"targets": []any{"0.0.0.0:8888"},
										},
									},
								},
							},
						},
					},
					"receiver_creator/discovery": map[string]any{
						"receivers": map[string]any{
							"smartagent/postgresql": map[string]any{
								"config": map[string]any{
									"connectionString": "sslmode=disable user={{.username}} password={{.password}}",
									"masterDBName":     "test_db",
									"params": map[string]any{
										"password": "<redacted>",
										"username": "<redacted>",
									},
									"type": "postgresql",
								},
								"resource_attributes": map[string]any{},
								"rule":                "type == \"container\" and any([name, image, command], {# matches \"(?i)postgres\"}) and not (command matches \"splunk.discovery\")",
							},
						},
						"watch_observers": []any{"docker_observer"},
					},
				},
				"service": map[string]any{
					"extensions": []any{"docker_observer"},
					"pipelines": map[string]any{
						"metrics": map[string]any{
							"exporters": []any{"otlp"},
							"receivers": []any{
								"prometheus/internal",
								"receiver_creator/discovery",
							},
						},
					},
				},
			}
			require.Equal(t, expectedEffective, cc.EffectiveConfig(t, 55554))

			rc, stdout, stderr := cc.Container.AssertExec(t, 5*time.Second, "tree", "/etc/otel/collector/config.d")
			require.Zero(t, rc, fmt.Sprintf("tree failed. stdout: %q, stderr: %q", stdout, stderr))
			require.Equal(t, expectedTree, stdout)
		})
	}
}
