package httpclient

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/signalfx/signalfx-agent/pkg/core/common/auth"
	"github.com/signalfx/signalfx-agent/pkg/utils/timeutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runServer(t *testing.T, h *HTTPConfig, cb func(host string)) {
	var err error
	req := require.New(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		// Optionally delay response return.
		if delay := request.URL.Query().Get("delay"); delay != "" {
			d, err := time.ParseDuration(delay)
			if err != nil {
				panic(err)
			}
			time.Sleep(d)
		}

		// So timeout test is more reliable.
		if h.Username != "" {
			if username, password, ok := request.BasicAuth(); !ok || username != h.Username || password != h.Password {
				writer.WriteHeader(http.StatusUnauthorized)
				_, _ = writer.Write([]byte("unauthorized"))
				return
			}
		}

		for configHeader, configValue := range h.HTTPHeaders {
			assert.Equal(t, configValue, request.Header.Get(configHeader))
		}

		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte("http body"))
	})
	server := &http.Server{
		Addr:    "localhost",
		Handler: mux,
	}

	listener, err := net.Listen("tcp", "localhost:0")
	req.NoError(err)

	serveReturn := make(chan error)
	defer func() {
		server.Close()
		err := <-serveReturn
		if err != nil && err != http.ErrServerClosed {
			t.Fatalf("failed stopping server: %s", err)
		}
	}()

	if h.UseHTTPS && h.ClientKeyPath != "" && h.ClientCertPath != "" {
		clientCA := x509.NewCertPool()
		err = auth.AugmentCertPoolFromCAFile(clientCA, "test-certs/client.pem")
		req.NoError(err)
		err = auth.AugmentCertPoolFromCAFile(clientCA, "test-certs/root.pem")
		req.NoError(err)
		server.TLSConfig = &tls.Config{
			ClientCAs:  clientCA,
			ClientAuth: tls.RequireAndVerifyClientCert,
		}
	}

	go func() {
		if h.UseHTTPS {
			serveReturn <- server.ServeTLS(listener, "test-certs/leaf.pem", "test-certs/leaf.key")
		} else {
			serveReturn <- server.Serve(listener)
		}
	}()

	// Need the dynamically allocated port but want to use localhost instead
	// of the returned 127.0.0.1.
	split := strings.Split(listener.Addr().String(), ":")
	req.Len(split, 2)

	cb("localhost:" + split[1])
}

func verify(t *testing.T, h *HTTPConfig) func(host string) {
	return func(host string) {
		req := require.New(t)
		client, err := h.Build()
		req.NoError(err)
		resp, err := client.Get((&url.URL{Scheme: h.Scheme(), Host: host}).String())
		req.NoError(err)
		defer func() {
			req.NoError(resp.Body.Close())
		}()
		req.Equal(http.StatusOK, resp.StatusCode)

		data, err := ioutil.ReadAll(resp.Body)
		req.NoError(err)

		req.Equal("http body", string(data))
	}
}

func TestHttpClientPlain(t *testing.T) {
	h := &HTTPConfig{UseHTTPS: false}
	runServer(t, h, verify(t, h))
}

func TestHttpsClientSkipVerify(t *testing.T) {
	h := &HTTPConfig{UseHTTPS: true, SkipVerify: true}
	runServer(t, h, verify(t, h))
}

func TestHttpsClientWithAuth(t *testing.T) {
	h := &HTTPConfig{
		UseHTTPS: true, CACertPath: "test-certs/leaf.pem", ClientCertPath: "test-certs/client.pem",
		ClientKeyPath: "test-certs/client.key"}
	runServer(t, h, verify(t, h))
}

func TestHttpBasicAuth(t *testing.T) {
	h := &HTTPConfig{Username: "bob", Password: "password"}
	runServer(t, h, verify(t, h))
}

func TestHttpTimeout(t *testing.T) {
	h := &HTTPConfig{
		HTTPTimeout: timeutil.Duration(1 * time.Nanosecond),
	}
	req := require.New(t)
	runServer(t, h, func(host string) {
		client, err := h.Build()
		req.NoError(err)
		// Need artificial delay to test timeout.
		u := url.URL{Scheme: h.Scheme(), Host: host, RawQuery: "delay=0.5s"}
		resp, err := client.Get(u.String())
		// To make linter happy.
		if err == nil {
			_ = resp.Body.Close()
		}
		req.Error(err)
		var uerr *url.Error
		req.True(errors.As(err, &uerr))
		req.True(uerr.Timeout())
	})
}

func TestHttpConfig_Scheme(t *testing.T) {
	type fields struct {
		UseHTTPS bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"https enabled", fields{UseHTTPS: true}, "https"},
		{"https disabled", fields{UseHTTPS: false}, "http"},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			h := &HTTPConfig{
				UseHTTPS: tt.fields.UseHTTPS,
			}
			if got := h.Scheme(); got != tt.want {
				t.Errorf("Scheme() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHttpHeaders(t *testing.T) {
	h := &HTTPConfig{
		HTTPHeaders: map[string]string{
			"HeaderOne": "ValueOne, ValueTwo",
			"HeaderTwo": "ValueThree",
		},
	}
	runServer(t, h, verify(t, h))
}
