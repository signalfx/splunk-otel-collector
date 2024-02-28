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
//go:build endtoend
// +build endtoend

package endtoend

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	sfxpb "github.com/signalfx/com_signalfx_metrics_protobuf/model"
	sfx "github.com/signalfx/signalfx-go"
	"github.com/signalfx/signalfx-go/metrics_metadata"
	"github.com/signalfx/signalfx-go/signalflow/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

// ./testdata/secrets.yaml
// To be broken out and finalized with additional suites.
// Must be exported for yaml unmarshalling to work.
type TestSecrets struct {
	DefaultToken   string `yaml:"default_token"`
	APIToken       string `yaml:"api_token"`
	IngestToken    string `yaml:"ingest_token"`
	IngestAPIToken string `yaml:"ingest_api_token"`
	APIUrl         string `yaml:"api_url"`
	IngestUrl      string `yaml:"ingest_url"`
	SignalFlowUrl  string `yaml:"signalflow_url"`
}

func secrets(t *testing.T) TestSecrets {
	content, err := ioutil.ReadFile(path.Join(".", "testdata", "secrets.yaml"))
	require.NoError(t, err)
	require.NotEmpty(t, content)

	var ts TestSecrets
	err = yaml.Unmarshal(content, &ts)

	require.NoError(t, err)
	require.NotEmpty(t, ts.DefaultToken, "'default_token' must be specified in secrets file")
	require.NotEmpty(t, ts.APIToken, "'api_token' must be specified in secrets file")
	require.NotEmpty(t, ts.IngestToken, "'ingest_token' must be specified in secrets file")
	require.NotEmpty(t, ts.IngestAPIToken, "'ingest_api_token' must be specified in secrets file")
	require.NotEmpty(t, ts.APIUrl, "'api_url' must be specified in secrets file")
	require.NotEmpty(t, ts.IngestUrl, "'ingest_url' must be specified in secrets file")
	require.NotEmpty(t, ts.SignalFlowUrl, "'signalflow_url' must be specified in secrets file")
	return ts
}

func TestIngestAuthScopeTokenGrantsRequiredMetricAndDimensionCapabilities(tt *testing.T) {
	ts := secrets(tt)
	runs := []struct {
		scope string
		token string
	}{
		{"ingest", ts.IngestToken},
		{"ingest/api", ts.IngestAPIToken},
	}
	for _, run := range runs {
		tt.Run(fmt.Sprintf("scope-%s", run.scope), func(t *testing.T) {
			tc := testutils.NewTestcase(t)
			defer tc.PrintLogsOnFailure()
			// Not used directly since we are sending to backend
			tc.ShutdownOTLPReceiverSink()

			_, stop := tc.Containers(postgresContainers()...)
			defer stop()

			os.Setenv("SPLUNK_ACCESS_TOKEN", run.token)
			os.Setenv("SPLUNK_API_URL", ts.APIUrl)
			os.Setenv("SPLUNK_INGEST_URL", ts.IngestUrl)

			collector, shutdown := tc.SplunkOtelCollector("postgres_config.yaml")
			defer shutdown()

			fmt.Println("sleeping 20s to give collector time to process and send metadata updates")
			time.Sleep(20 * time.Second)

			client, err := sfx.NewClient(ts.DefaultToken, sfx.APIUrl(ts.APIUrl))
			require.NoError(t, err)
			require.NotNil(t, client)

			mtses := assertFoundMetricTimeSeries(
				t, client, fmt.Sprintf("metric:postgres_queries_total_time AND testid:%s", tc.ID),
			)
			assertQueryIdDimensionsWithQueryProperties(tc, client, mtses)

			mtses = assertFoundMetricTimeSeries(
				t, client, fmt.Sprintf("metric:postgres_queries_calls AND testid:%s", tc.ID),
			)
			assertQueryIdDimensionsWithQueryProperties(tc, client, mtses)

			assertLogsDontContain401(tc, collector, run.scope)
		})
	}
}

func TestIngestAuthScopeTokenGrantsRequiredEventCapabilities(tt *testing.T) {
	ts := secrets(tt)
	runs := []struct {
		scope string
		token string
	}{
		{"ingest", ts.IngestToken},
		{"ingest/api", ts.IngestAPIToken},
	}
	for _, run := range runs {
		tt.Run(fmt.Sprintf("scope-%s", run.scope), func(t *testing.T) {
			tc := testutils.NewTestcase(t)
			defer tc.PrintLogsOnFailure()
			// Not used directly since we are sending to backend
			tc.ShutdownOTLPReceiverSink()

			os.Setenv("SPLUNK_ACCESS_TOKEN", run.token)
			os.Setenv("SPLUNK_API_URL", ts.APIUrl)
			os.Setenv("SPLUNK_INGEST_URL", ts.IngestUrl)

			collector, shutdown := tc.SplunkOtelCollector("event_forwarder_config.yaml")
			defer shutdown()

			client, err := signalflow.NewClient(
				signalflow.AccessToken(ts.DefaultToken),
				signalflow.StreamURL(ts.SignalFlowUrl),
				signalflow.OnError(func(err error) {
					require.NoError(t, err)
				}),
			)
			require.NoError(t, err)
			require.NotNil(t, client)

			eventType := fmt.Sprintf("testevent%s", tc.ID)
			program := fmt.Sprintf("events(eventType=%q).publish()", eventType)

			comp, err := client.Execute(context.Background(), &signalflow.ExecuteRequest{Program: program})
			require.NoError(t, err)
			require.NotNil(t, comp)
			defer comp.Stop(context.Background())

			fmt.Println("sleeping 10s before sending event to Collector to allow computation to begin")
			time.Sleep(10 * time.Second)

			sendEvents(tc.TB, eventType)

			done := make(chan struct{})
			go func() {
				expected := 10
				for event := range comp.Events() {
					raw := event.RawData()
					require.Contains(t, raw, "metadata")
					// added by configured resource processor
					require.Contains(t, raw["metadata"], "testid")
					metadata, ok := raw["metadata"].(map[string]any)
					require.True(t, ok)
					require.Equal(t, tc.ID, metadata["testid"])
					expected--
					if expected == 0 {
						close(done)
						return
					}
				}
			}()

			err = comp.Err()
			require.NoError(t, err)

			select {
			case <-done:
			case <-time.After(60 * time.Second):
				require.Fail(t, "expected event never received")
			}

			assertLogsDontContain401(tc, collector, run.scope)
		})
	}
}

func TestAPIAuthScopeTokenDoesntGrantRequiredMetricAndDimensionCapabilities(t *testing.T) {
	ts := secrets(t)
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	tc.ShutdownOTLPReceiverSink()

	_, stop := tc.Containers(postgresContainers()...)
	defer stop()

	os.Setenv("SPLUNK_ACCESS_TOKEN", ts.APIToken)
	os.Setenv("SPLUNK_API_URL", ts.APIUrl)
	os.Setenv("SPLUNK_INGEST_URL", ts.IngestUrl)

	collector, shutdown := tc.SplunkOtelCollector("postgres_config.yaml")
	defer shutdown()

	client, err := sfx.NewClient(ts.DefaultToken, sfx.APIUrl(ts.APIUrl))
	require.NoError(t, err)
	require.NotNil(t, client)

	assertNoFoundMetricTimeSeries(t, client, fmt.Sprintf("metric:postgres_queries_total_time AND testid:%s", tc.ID))
	assertNoFoundMetricTimeSeries(t, client, fmt.Sprintf("metric:postgres_queries_calls AND testid:%s", tc.ID))

	assertLogsContain401(tc, collector, "api")
}

func TestAPIAuthScopeTokenDoesntGrantRequiredEventCapabilities(t *testing.T) {
	ts := secrets(t)
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	tc.ShutdownOTLPReceiverSink()

	os.Setenv("SPLUNK_ACCESS_TOKEN", ts.APIToken)
	os.Setenv("SPLUNK_API_URL", ts.APIUrl)
	os.Setenv("SPLUNK_INGEST_URL", ts.IngestUrl)

	collector, shutdown := tc.SplunkOtelCollector("event_forwarder_config.yaml")
	defer shutdown()

	sflow, err := signalflow.NewClient(
		signalflow.AccessToken(ts.DefaultToken),
		signalflow.StreamURL(ts.SignalFlowUrl),
		signalflow.OnError(func(err error) {
			require.NoError(t, err)
		}),
	)

	require.NoError(t, err)
	require.NotNil(t, sflow)

	eventType := fmt.Sprintf("testevent%s", tc.ID)
	program := fmt.Sprintf("events(eventType=%q).publish()", eventType)

	comp, err := sflow.Execute(context.Background(), &signalflow.ExecuteRequest{Program: program})
	require.NoError(t, err)
	require.NotNil(t, comp)
	defer comp.Stop(context.Background())

	fmt.Println("sleeping 10s before sending event to Collector to allow computation to begin")
	time.Sleep(10 * time.Second)

	sendEvents(tc.TB, eventType)

	done := make(chan struct{})
	go func() {
		for event := range comp.Events() {
			close(done)
			require.Fail(t, "should never have received with api token event: %v", event)
			return
		}
	}()

	err = comp.Err()
	require.NoError(t, err)

	select {
	case <-done:
		require.Fail(t, "received unexpected event")
	case <-time.After(60 * time.Second):
	}

	assertLogsContain401(tc, collector, "api")
}

// Postgres monitor uses a query property on queryid dimension which we can examine to confirm token permissions
func postgresContainers() []testutils.Container {
	postgresServer := testutils.NewContainer().WithContext(
		path.Join("..", "receivers", "smartagent", "postgresql", "testdata", "server"),
	).WithEnv(
		map[string]string{"POSTGRES_DB": "test_db", "POSTGRES_USER": "postgres", "POSTGRES_PASSWORD": "postgres"},
	).WithExposedPorts("5432:5432").WithName("postgres-server").WithNetworks(
		"postgres",
	).WillWaitForPorts("5432").WillWaitForLogs("database system is ready to accept connections")

	postgresClient := testutils.NewContainer().WithContext(
		path.Join("..", "receivers", "smartagent", "postgresql", "testdata", "client"),
	).WithEnv(
		map[string]string{"POSTGRES_SERVER": "postgres-server"},
	).WithName("postgres-client").WithNetworks("postgres").WillWaitForLogs("Beginning psql requests")
	return []testutils.Container{postgresServer, postgresClient}
}

func assertFoundMetricTimeSeries(t *testing.T, sfxClient *sfx.Client, query string) metrics_metadata.MetricTimeSeriesRetrieveResponseModel {
	var foundMts *metrics_metadata.MetricTimeSeriesRetrieveResponseModel
	require.Eventually(t, func() bool {
		var err error
		foundMts, err = sfxClient.SearchMetricTimeSeries(context.Background(), query, "", 100, 0)
		require.NoError(t, err)
		require.NotNil(t, foundMts)
		return foundMts.Count > 0 && len(foundMts.Results) > 0
	}, 30*time.Second, time.Second)
	return *foundMts
}

func assertNoFoundMetricTimeSeries(t *testing.T, sfxClient *sfx.Client, query string) {
	require.Never(t, func() bool {
		foundMts, err := sfxClient.SearchMetricTimeSeries(context.Background(), query, "", 100, 0)
		require.NoError(t, err)
		return foundMts.Count > 0 && len(foundMts.Results) > 0
	}, 30*time.Second, time.Second)
}

func assertQueryIdDimensionsWithQueryProperties(tc *testutils.Testcase, sfxClient *sfx.Client, mtses metrics_metadata.MetricTimeSeriesRetrieveResponseModel) {
	for _, mts := range mtses.Results {
		require.Equal(tc, mts.Dimensions["testid"], tc.ID)
		assert.NotContains(tc, mts.Dimensions, "query")
		assert.Contains(tc, mts.CustomProperties, "query")
		assert.Equal(tc, mts.Dimensions["queryid"], mts.CustomProperties["queryid"])
		queryId, err := sfxClient.GetDimension(context.Background(), "queryid", mts.Dimensions["queryid"].(string))
		require.NoError(tc, err)
		require.NotNil(tc, queryId)
		require.Contains(tc, queryId.CustomProperties, "query")

		command := strings.ToLower(strings.Split(queryId.CustomProperties["query"], " ")[0])
		require.True(tc, func() bool {
			commands := []string{"create", "delete", "drop", "grant", "insert", "select", "update", "with"}
			for _, cmd := range commands {
				if command == cmd {
					return true
				}
			}
			return false
		}(), "%s not in expected commands", command)
	}
}

func sendEvents(t testing.TB, eventType string) {
	propVal := "a test event"
	description := sfxpb.Property{Key: "description", Value: &sfxpb.PropertyValue{StrValue: &propVal}}

	alert := sfxpb.EventCategory_USER_DEFINED

	// I have found sending several events to be necessary for detecting in a fresh signalflow job
	for i := 0; i < 10; i++ {
		dim := sfxpb.Dimension{Key: "dim_one", Value: fmt.Sprintf("%d", i)}
		event := sfxpb.Event{
			EventType:  eventType,
			Dimensions: []*sfxpb.Dimension{&dim},
			Properties: []*sfxpb.Property{&description},
			Category:   &alert,
			Timestamp:  time.Now().UnixNano() / int64(time.Millisecond), // convert to ms (required for signalflow reporting)
		}

		msg := sfxpb.EventUploadMessage{Events: []*sfxpb.Event{&event}}

		eventBody, err := msg.Marshal()
		require.NoError(t, err)

		req, err := http.NewRequest("POST", "http://localhost:8006/v2/event", bytes.NewBuffer(eventBody))
		req.Header.Set("Content-Type", "application/x-protobuf")

		httpClient := &http.Client{}
		resp, err := httpClient.Do(req)
		io.Copy(io.Discard, resp.Body)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		time.Sleep(500 * time.Millisecond)
	}
}

func logsContain401(tc *testutils.Testcase, collector testutils.Collector) bool {
	errorMsg := "HTTP 401"
	errorMsg2 := "HTTP/2.0 401"

	// There is a bug in testcontainers-go so we can't rely on the
	// LogProducer until resolved: https://github.com/testcontainers/testcontainers-go/pull/323
	if ctr, ok := collector.(*testutils.CollectorContainer); ok {
		logs, err := ctr.Container.Logs(context.Background())
		require.NoError(tc, err)
		buf := new(strings.Builder)
		_, err = io.Copy(buf, logs)
		require.NoError(tc, err)
		errContent := buf.String()
		return strings.Contains(errContent, errorMsg) || strings.Contains(errContent, errorMsg2)
	}

	for _, statement := range tc.ObservedLogs.All() {
		if strings.Contains(fmt.Sprintf("%v", statement), errorMsg) {
			return true
		}
	}
	return false
}

func assertLogsDontContain401(tc *testutils.Testcase, collector testutils.Collector, scope string) {
	require.False(tc, logsContain401(tc, collector), "undesired 401 error observed for token with %s auth scope", scope)
}

func assertLogsContain401(tc *testutils.Testcase, collector testutils.Collector, scope string) {
	require.True(tc, logsContain401(tc, collector), "desired 401 error not observed for token with %s auth scope", scope)
}
