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
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/Masterminds/semver"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"

	"github.com/signalfx/splunk-otel-collector/tests/internal/version"
	"github.com/signalfx/splunk-otel-collector/tests/testutils"
	"github.com/signalfx/splunk-otel-collector/tests/testutils/kubeutils"
)

const (
	chartPathEnvVar    = "SPLUNK_COLLECTOR_HELM_CHART_PATH"
	chartVersionEnvVar = "SPLUNK_COLLECTOR_HELM_CHART_VERSION"
	helmRepoTmpl       = "https://github.com/signalfx/splunk-otel-collector-chart/releases/download/splunk-otel-collector-%s/splunk-otel-collector-%s.tgz"
)

func getHelmChartPath(t testing.TB) string {
	if chartPath := strings.TrimSpace(os.Getenv(chartPathEnvVar)); chartPath != "" {
		return chartPath
	}

	var ver string
	if ver = strings.TrimSpace(os.Getenv(chartVersionEnvVar)); ver == "" {
		if matches := regexp.MustCompile("v(?P<version>[\\d.]+)(-.*)?").FindStringSubmatch(version.Version); len(matches) == 3 {
			v, err := semver.NewVersion(matches[1])
			require.NoError(t, err)
			v, err = semver.NewVersion(fmt.Sprintf("%d.%d.0", v.Major(), v.Minor()-1))
			require.NoError(t, err)
			ver = v.String()
		}
	}

	require.NotEmpty(t, ver, "Must set %q or %q to determine chart path if not valid version.Version", chartPathEnvVar, chartVersionEnvVar)
	return fmt.Sprintf(helmRepoTmpl, ver, ver)
}

// TestHelmChartMetricsHappyPath ensures that the default metrics (except for env-required filters)
// for a helm chart installation are exported. The target chart is determined from first hit of
// 1. SPLUNK_COLLECTOR_HELM_CHART_PATH env var value
// 2. SPLUNK_COLLECTOR_HELM_CHART_VERSION env var set in helmRepoTmpl
// 3. parsed version.Version (with decremented minor to check previous release compatibility) set in helmRepoTmpl
func TestHelmChartMetricsHappyPath(t *testing.T) {
	testutils.SkipIfNotContainerTest(t)
	if testutils.CollectorImageIsForArm(t) {
		t.Skip("Apparent metric loss on qemu. Deferring.")
	}
	helmChartPath := getHelmChartPath(t)
	tc := testutils.NewTestcase(t, testutils.OTLPReceiverSinkAllInterfaces)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	cluster := kubeutils.NewKindCluster(tc)
	defer cluster.Teardown()
	cluster.Create()

	cluster.LoadLocalCollectorImageIfNecessary()

	defer kubeutils.NewOTLPSinkDeployment(cluster).Teardown()

	helm := kubeutils.Helm(cluster, func(settings *cli.EnvSettings) {
		settings.SetNamespace("monitoring")
	})

	img := strings.Split(testutils.GetCollectorImage(), ":")
	image := img[0]
	tag := "latest"
	if len(img) == 2 {
		tag = img[1]
	}

	values := fmt.Sprintf(`clusterName: test-cluster
environment: test-cluster
image:
  otelcol:
    repository: %s
    tag: %s

splunkObservability:
  realm: noop
  accessToken: splunk-o11y-token
  apiUrl: http://otlp-sink.testing.svc.cluster.local:26060
  ingestUrl: http://otlp-sink.testing.svc.cluster.local:29943
  metricsEnabled: true
  tracesEnabled: false
  logsEnabled: false

agent:
  config:
    receivers:
      kubeletstats:
        insecure_skip_verify: true
        # current gh action runner
        # doesn't have working kubelet cadvisor
        # w/ systemd or cgroupfs cgroup driver so avoiding these
        # https://github.com/kubernetes/kubernetes/issues/103366#issuecomment-887247862
        metrics:
          k8s.pod.network.errors:
            enabled: false
          k8s.pod.network.io:
            enabled: false
gateway:
  enabled: true
  replicaCount: 1
  resources:
    limits:
      cpu: 200m
      memory: 128Mi
`, image, tag)

	release, err := helm.Install(
		helmChartPath, values,
		func(install *action.Install) {
			install.CreateNamespace = true
			install.Namespace = "monitoring"
		},
	)
	require.NoError(t, err)

	for _, pod := range []string{"agent-.*", "k8s-cluster-receiver-.*", ".{16}"} {
		cluster.WaitForPods(fmt.Sprintf("%s-%s", release.Name, pod), "monitoring", 2*time.Minute)
	}

	tc.OTLPReceiverSink.AssertAllMetricsReceived(tc, *tc.ResourceMetrics("all.yaml"), 30*time.Second)
}
