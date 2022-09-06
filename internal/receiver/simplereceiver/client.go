// This is somewhat unique to this type of receiver, the logic to create and dispatch requests
// is defined here.

package simplereceiver

import (
	"net/http"
	"net/url"

	"go.opentelemetry.io/collector/component"
)

// default client which will
type defaultClient struct {
    client *http.Client
    endpoint *url.URL
}

func newDefaultClient() defaultClient {}

func (c defaultClient) makeRequest() {}


