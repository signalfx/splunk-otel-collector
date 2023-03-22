package auth

import (
	"fmt"
	"net/http"
)

// An http transport that injects an OAuth bearer token onto each request
type TransportWithToken struct {
	http.RoundTripper
	Token string
}

// Override the only method that the client actually calls on the transport to
// do the request.
func (t *TransportWithToken) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", t.Token))
	return t.RoundTripper.RoundTrip(req)
}
