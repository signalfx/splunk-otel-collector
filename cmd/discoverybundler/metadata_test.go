// Copyright  Splunk, Inc.
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

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestRabbitMQDiscoveryMetadataUsesManagementEndpoint(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("metadata", "receivers", "rabbitmq.yaml"))
	require.NoError(t, err)

	var metadata receiverMetadata
	require.NoError(t, yaml.Unmarshal(data, &metadata))

	tmpl, err := template.New("rabbitmq").Funcs(funcMap(metadata.ReceiverID)).Parse(metadata.PropertiesTmpl)
	require.NoError(t, err)

	var rendered bytes.Buffer
	require.NoError(t, tmpl.Execute(&rendered, nil))

	var properties struct {
		Rule   map[string]string `yaml:"rule"`
		Config struct {
			Default struct {
				Endpoint string `yaml:"endpoint"`
			} `yaml:"default"`
		} `yaml:"config"`
	}
	require.NoError(t, yaml.Unmarshal(rendered.Bytes(), &properties))

	require.Equal(t, "http://`endpoint`", properties.Config.Default.Endpoint)
	require.Equal(t, `type == "container" and port == 15672 and any([name, image, command], {# matches "(?i)rabbitmq.*"}) and not (command matches "splunk.discovery")`, properties.Rule["docker_observer"])
	require.Equal(t, `type == "hostport" and port == 15672 and command matches "(?i)rabbitmq.*" and not (command matches "splunk.discovery")`, properties.Rule["host_observer"])
	require.Equal(t, `type == "port" and port == 15672 and pod.name matches "(?i)rabbitmq.*"`, properties.Rule["k8s_observer"])
}
