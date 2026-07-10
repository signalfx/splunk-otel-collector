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

type AgentTemplateDestination struct {
	ConfigDestinationFilePath  string
	SplunkAPIDestinationURL    string
	SplunkAPIPort              string
	SplunkDestinationURL       string
	SplunkIngestPort           string
	OpAmpPort                  string
	OTLPEntitiesExporterConfig string
	OTLPEntitiesExporterName   string
	OTLPExporterConfig         string
	OTLPExporterName           string
}

func main() {
	agentConfigs := []AgentTemplateDestination{
		{
			ConfigDestinationFilePath: filepath.Join("..", "otelcol", "config", "collector", "agent_config.yaml"),
			SplunkAPIDestinationURL:   "${SPLUNK_API_URL}",
			SplunkDestinationURL:      "${SPLUNK_INGEST_URL}",
			OTLPEntitiesExporterConfig: `logs_endpoint: "${SPLUNK_INGEST_URL}/v3/event"
    headers:
      "X-SF-Token": "${SPLUNK_ACCESS_TOKEN}"
    auth:
      authenticator: headers_setter`,
			OTLPEntitiesExporterName: "otlp_http/entities",
			OTLPExporterConfig: `traces_endpoint: "${SPLUNK_INGEST_URL}/v2/trace/otlp"
    headers:
      "X-SF-Token": "${SPLUNK_ACCESS_TOKEN}"
    auth:
      authenticator: headers_setter`,
			OTLPExporterName: "otlp_http",
		},
		{
			ConfigDestinationFilePath: filepath.Join("..", "otelcol", "config", "collector", "agent_to_gateway_config.yaml"),
			SplunkAPIDestinationURL:   "${SPLUNK_GATEWAY_URL}",
			SplunkAPIPort:             "6060",
			SplunkDestinationURL:      "${SPLUNK_GATEWAY_URL}",
			SplunkIngestPort:          "9943",
			OpAmpPort:                 "4320",
			OTLPExporterConfig: `endpoint: "${SPLUNK_GATEWAY_URL}:4317"
    tls:
      insecure: true
    auth:
      authenticator: headers_setter`,
			OTLPExporterName: "otlp_grpc/gateway",
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
