// Copyright Splunk, Inc.
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

package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

type TelemetryDestination struct {
	URL  string
	Port string
}

type ComponentConfig struct {
	Name     string
	Contents string
}

// TODO: How to designate some options as required, others as optional?
// When adding a new destination, how will a developer know which fields
// must be set, and which can be left blank?
type AgentTemplateDestination struct {
	LogsExporter              ComponentConfig
	OTLPEntitiesExporter      ComponentConfig
	OTLPGenericExporter       ComponentConfig
	ProfilingExporter         ComponentConfig
	OpAmp                     TelemetryDestination
	SplunkAPI                 TelemetryDestination
	SplunkIngest              TelemetryDestination
	ConfigDestinationFilePath string
}

func main() {
	agentConfigs := []AgentTemplateDestination{
		{
			LogsExporter: ComponentConfig{
				Name: "splunk_hec",
				Contents: `token: "${SPLUNK_HEC_TOKEN}"
    endpoint: "${SPLUNK_HEC_URL}"
    source: "otel"
    sourcetype: "otel"
    profiling_data_enabled: false`,
			},
			OTLPEntitiesExporter: ComponentConfig{
				Name: "otlp_http/entities",
				Contents: `logs_endpoint: "${SPLUNK_INGEST_URL}/v3/event"
    headers:
      "X-SF-Token": "${SPLUNK_ACCESS_TOKEN}"
    auth:
      authenticator: headers_setter`,
			},
			OTLPGenericExporter: ComponentConfig{
				Name: "otlp_http",
				Contents: `traces_endpoint: "${SPLUNK_INGEST_URL}/v2/trace/otlp"
    headers:
      "X-SF-Token": "${SPLUNK_ACCESS_TOKEN}"
    auth:
      authenticator: headers_setter`,
			},
			ProfilingExporter: ComponentConfig{
				Name: "splunk_hec/profiling",
				Contents: `token: "${SPLUNK_ACCESS_TOKEN}"
    endpoint: "${SPLUNK_INGEST_URL}/v1/log"
    log_data_enabled: false`,
			},
			SplunkAPI: TelemetryDestination{
				URL: "${SPLUNK_API_URL}",
			},
			SplunkIngest: TelemetryDestination{
				URL: "${SPLUNK_INGEST_URL}",
			},
			ConfigDestinationFilePath: filepath.Join("..", "otelcol", "config", "collector", "agent_config.yaml"),
		},
		{
			SplunkAPI: TelemetryDestination{
				URL:  "${SPLUNK_GATEWAY_URL}",
				Port: "6060",
			},
			SplunkIngest: TelemetryDestination{
				URL:  "${SPLUNK_GATEWAY_URL}",
				Port: "9943",
			},
			OpAmp: TelemetryDestination{
				Port: "4320",
			},
			OTLPGenericExporter: ComponentConfig{
				Name: "otlp_grpc/gateway",
				Contents: `endpoint: "${SPLUNK_GATEWAY_URL}:4317"
    tls:
      insecure: true
    auth:
      authenticator: headers_setter`,
			},
			ConfigDestinationFilePath: filepath.Join("..", "otelcol", "config", "collector", "agent_to_gateway_config.yaml"),
		},
	}

	tmpl, err := template.ParseFiles(filepath.Join("config_templates", "agent_config_source.yaml.tmpl"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing templates: %v\n", err)
		os.Exit(1)
	}

	for _, agentConfig := range agentConfigs {
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, agentConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error executing template: %v\n", err)
			os.Exit(1)
		}

		err = os.WriteFile(agentConfig.ConfigDestinationFilePath, buf.Bytes(), 0o444)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing config: %v\n", err)
			os.Exit(1)
		}
	}
}
