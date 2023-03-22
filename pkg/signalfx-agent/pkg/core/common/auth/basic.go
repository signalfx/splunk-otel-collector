package auth

import (
	"net/http"
)

// An http transport that injects basic auth into each request
type TransportWithBasicAuth struct {
	http.RoundTripper
	Username string
	Password string
}

// Override the only method that the client actually calls on the transport to
// do the request.
func (t *TransportWithBasicAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(t.Username, t.Password)
	return t.RoundTripper.RoundTrip(req)
}
