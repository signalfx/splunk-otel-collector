// test code for client.go

package githubmetricsreceiver

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/confighttp"
)

func testClientCreation(t *testing.T) {
    _, err := newDefaultClient(componenttest.NewNopTelemetrySettings(), Config{
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Endpoint: "http://api.github.com",
		},
	}, componenttest.NewNopHost())
    assert.Equal(t, err, nil)
}

func testGetRepoChanges(t *testing.T) {
    payload, _ := os.ReadFile("testdata/commit_activity_test_data.json")

    commstats := []commitStats{}
    err := json.Unmarshal(payload, &commstats)

    assert.Equal(t, err, nil)
    assert.Equal(t, commstats[len(commstats)-1].WeekStamp, 1662249600)
}

func testGetCommitStats(t *testing.T) {
    payload, _ := os.ReadFile("testdata/code_frequency_test_data.json")

    comAct, err := newCommitActivity(payload)

    assert.Equal(t, err, nil)
    assert.Equal(t, comAct.WeekStamp, 1662249600)
}
