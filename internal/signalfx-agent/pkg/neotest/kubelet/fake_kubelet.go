package kubelet

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
)

// FakeKubelet is a mock of the ingest server.  Holds all of the received
// datapoints for later inspection
type FakeKubelet struct {
	server  *httptest.Server
	PodJSON []byte
}

// NewFakeKubelet creates a new instance of FakeKubelet but does not start
// the server
func NewFakeKubelet() *FakeKubelet {
	return &FakeKubelet{}
}

// Start creates and starts the mock HTTP server
func (f *FakeKubelet) Start() {
	f.server = httptest.NewUnstartedServer(f)
	f.server.Start()
}

// Close stops the mock HTTP server
func (f *FakeKubelet) Close() {
	f.server.Close()
}

// URL is the of the mock server to point your objects under test to
func (f *FakeKubelet) URL() *url.URL {
	url, err := url.Parse(f.server.URL)
	if err != nil {
		panic("Bad URL " + url.String())
	}
	return url
}

// ServeHTTP handles a single request
func (f *FakeKubelet) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	//contents, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if r.URL.Path == "/pods" {
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write(f.PodJSON)
	} else {
		rw.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(rw, "Not found")
	}
}
