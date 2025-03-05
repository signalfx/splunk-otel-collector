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
	assert.Equal(t, "_8ecddlaO_uQyTlmLigzukO3yo5M3b68gl8nge72jA7Dam3B2iPlzUlLUh7xucJnzs44VR0NzuvH9UHsl1xr0B6vNjJ9bGUC3HlCQnXf_94ikUzryC", input.SessionKey, "Expected SessionKey to be '_8ecddlaO_uQyTlmLigzukO3yo5M3b68gl8nge72jA7Dam3B2iPlzUlLUh7xucJnzs44VR0NzuvH9UHsl1xr0B6vNjJ9bGUC3HlCQnXf_94ikUzryC'")
	assert.Equal(t, `C:\Program Files\Splunk\var\lib\splunk\modinputs\Splunk_TA_otel`, input.CheckpointDir, "Expected CheckpointDir to be 'C:\\Program Files\\Splunk\\var\\lib\\splunk\\modinputs\\Splunk_TA_otel'")
	require.Len(t, input.Configuration.Stanza, 1, "Expected 1 Stanza")
	stanza := input.Configuration.Stanza[0]
	assert.Equal(t, "Splunk_TA_otel://Splunk_TA_otel", stanza.Name, "Expected Stanza Name to be 'Splunk_TA_otel://Splunk_TA_otel'")
	assert.Equal(t, "Splunk_TA_otel", stanza.App, "Expected Stanza App to be 'Splunk_TA_otel'")
	assert.Len(t, stanza.Param, 16, "Expected 16 Params")
}

func TestUnmarshalInputXML_InvalidXML(t *testing.T) {
	invalidXMLData := `<input><server_host>R91395DV</server_host>`

	_, err := UnmarshalInputXML([]byte(invalidXMLData))
	require.Error(t, err, "Expected error")
}
