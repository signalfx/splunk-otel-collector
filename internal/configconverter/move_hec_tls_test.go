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

package configconverter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMoveHecTLS(t *testing.T) {
	cp := &converterProvider{
		wrapped:     &fileParserProvider{fileName: "testdata/hec-tls.yaml"},
		cfgMapFuncs: []CfgMapFunc{MoveHecTLS},
	}
	cfgMap, err := cp.Get(context.Background())
	require.NoError(t, err)
	assert.False(t, cfgMap.IsSet("exporters::splunk_hec::ca_file"))
	assert.True(t, true, cfgMap.Get("exporters::splunk_hec::tls::insecure_skip_verify"))
	assert.Equal(t, "my-ca-file-1", cfgMap.Get("exporters::splunk_hec::tls::ca_file"))
	assert.Equal(t, "my-cert-file-1", cfgMap.Get("exporters::splunk_hec::tls::cert_file"))
	assert.Equal(t, "my-key-file-1", cfgMap.Get("exporters::splunk_hec::tls::key_file"))

	assert.False(t, cfgMap.IsSet("exporters::splunk_hec/allsettings::ca_file"))
	assert.True(t, true, cfgMap.Get("exporters::splunk_hec/allsettings::tls::insecure_skip_verify"))
	assert.Equal(t, "my-ca-file-2", cfgMap.Get("exporters::splunk_hec/allsettings::tls::ca_file"))
	assert.Equal(t, "my-cert-file-2", cfgMap.Get("exporters::splunk_hec/allsettings::tls::cert_file"))
	assert.Equal(t, "my-key-file-2", cfgMap.Get("exporters::splunk_hec/allsettings::tls::key_file"))
}
