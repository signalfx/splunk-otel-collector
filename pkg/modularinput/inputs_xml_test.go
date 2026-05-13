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

package modularinput

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalInputXML(t *testing.T) {
	xmlPath := "testdata/Splunk_TA_otel.inputs.xml"
	xmlData, err := os.ReadFile(xmlPath)
	require.NoError(t, err, "Failed to read file %s", xmlPath)

	input, err := UnmarshalInputXML(xmlData)
	require.NoError(t, err, "Expected no error")

	assert.Equal(t, "R91395DV", input.ServerHost, "Expected ServerHost to be 'R91395DV'")
	assert.Equal(t, "https://127.0.0.1:8089", input.ServerURI, "Expected ServerURI to be 'https://127.0.0.1:8089'")
	assert.Equal(t, "_fake_session_key", input.SessionKey, "Expected SessionKey to be '_fake_session_key'")
	assert.Equal(t, `C:\Program Files\Splunk\var\lib\splunk\modinputs\Splunk_TA_otel`, input.CheckpointDir, "Expected CheckpointDir to be 'C:\\Program Files\\Splunk\\var\\lib\\splunk\\modinputs\\Splunk_TA_otel'")
	require.Len(t, input.Configuration.Stanza, 1, "Expected 1 Stanza")
	stanza := input.Configuration.Stanza[0]
	assert.Equal(t, "Splunk_TA_otel://Splunk_TA_otel", stanza.Name, "Expected Stanza Name to be 'Splunk_TA_otel://Splunk_TA_otel'")
	assert.Equal(t, "Splunk_TA_otel", stanza.App, "Expected Stanza App to be 'Splunk_TA_otel'")
	assert.Len(t, stanza.Param, 15)
}

func TestUnmarshalInputXML_InvalidXML(t *testing.T) {
	invalidXMLData := `<input><server_host>R91395DV</server_host>`

	_, err := UnmarshalInputXML([]byte(invalidXMLData))
	require.Error(t, err, "Expected error")
}
