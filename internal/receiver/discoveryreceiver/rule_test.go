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

// The code is copied from
// https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/c07d1e622c59ed013e6a02f96a5cc556513263da/receiver/receivercreator/rules.go
// with minimal changes. Once the discovery receiver upstreamed, this code can be reused.

package discoveryreceiver

import (
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuleEval(t *testing.T) {
	type args struct {
		endpoint observer.Endpoint
		ruleStr  string
	}
	tests := []struct {
		args    args
		name    string
		want    bool
		wantErr bool
	}{
		{
			name:    "basic port",
			args:    args{portEndpoint, `type == "port" && name == "port.name" && pod.labels["label.one"] == "value.one"`},
			want:    true,
			wantErr: false,
		},
		{
			name:    "basic hostport",
			args:    args{hostportEndpoint, `type == "hostport" && port == 1 && process_name == "process.name"`},
			want:    true,
			wantErr: false,
		},
		{
			name:    "basic pod",
			args:    args{podEndpoint, `type == "pod" && labels["label.one"] == "value.one"`},
			want:    true,
			wantErr: false,
		},
		{
			name:    "annotations",
			args:    args{podEndpoint, `type == "pod" && annotations["annotation.one"] == "value.one"`},
			want:    true,
			wantErr: false,
		},
		{
			name:    "basic container",
			args:    args{containerEndpoint, `type == "container" && labels["label.one"] == "value.one"`},
			want:    true,
			wantErr: false,
		},
		{
			name:    "basic k8s.node",
			args:    args{k8sNodeEndpoint, `type == "k8s.node" && kubelet_endpoint_port == 1`},
			want:    true,
			wantErr: false,
		},
		{
			name:    "basic pod.container",
			args:    args{podContainerEndpoint, `type == "pod.container" && container_image matches "redis"`},
			want:    true,
			wantErr: false,
		},
		{
			name:    "relocated type builtin",
			args:    args{k8sNodeEndpoint, `type == "k8s.node" && typeOf("some string") == "string"`},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newRule(tt.args.ruleStr)
			require.NoError(t, err)
			require.NotNil(t, got)

			env, err := tt.args.endpoint.Env()
			require.NoError(t, err)

			match, err := got.eval(env)

			if (err != nil) != tt.wantErr {
				t.Errorf("eval() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.want, match, "expected eval to return %v but returned %v", tt.want, match)
		})
	}
}

func TestNewRule(t *testing.T) {
	type args struct {
		ruleStr string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty rule", args{""}, true},
		{"does not startMetrics with type", args{"port == 1234"}, true},
		{"invalid syntax", args{"port =="}, true},
		{"valid port", args{`type == "port" && port_name == "http"`}, false},
		{"valid pod", args{`type=="pod" && port_name == "http"`}, false},
		{"valid hostport", args{`type == "hostport" && port_name == "http"`}, false},
		{"valid container", args{`type == "container" && port == 8080`}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newRule(tt.args.ruleStr)
			if err == nil {
				assert.NotNil(t, got, "expected rule to be created when there was no error")
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("newRule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
