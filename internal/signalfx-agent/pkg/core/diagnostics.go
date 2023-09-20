package core

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	// Import for side-effect of registering http handler
	_ "net/http/pprof"
)

// VersionLine should be populated by the startup logic to contain version
// information that can be reported in diagnostics.
// nolint: gochecknoglobals
var VersionLine string

// Serves the diagnostic status on the specified path

func readStatusInfo(host string, port uint16, section string) ([]byte, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/?section=%s", host, port, section))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func streamDatapoints(host string, port uint16, metric string, dims string) (io.ReadCloser, error) {
	c := http.Client{
		Timeout: 0,
	}
	qs := url.Values{}
	qs.Set("metric", metric)
	qs.Set("dims", dims)
	resp, err := c.Get(fmt.Sprintf("http://%s:%d/tap-dps?%s", host, port, qs.Encode())) // nolint:bodyclose
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}
