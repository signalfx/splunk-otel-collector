package httpclient

import (
	"crypto/tls"
	"net/http"

	"github.com/signalfx/signalfx-agent/pkg/utils/timeutil"

	"github.com/signalfx/signalfx-agent/pkg/core/common/auth"
	"github.com/signalfx/signalfx-agent/pkg/core/common/constants"
)

// HTTPConfig can be embedded inside a monitor config.
type HTTPConfig struct {
	// HTTP timeout duration for both read and writes. This should be a
	// duration string that is accepted by https://golang.org/pkg/time/#ParseDuration
	HTTPTimeout timeutil.Duration `yaml:"httpTimeout" default:"10s"`

	// Basic Auth username to use on each request, if any.
	Username string `yaml:"username"`
	// Basic Auth password to use on each request, if any.
	Password string `yaml:"password" neverLog:"true"`

	// If true, the agent will connect to the server using HTTPS instead of plain HTTP.
	UseHTTPS bool `yaml:"useHTTPS"`

	// A map of HTTP header names to values. Comma separated multiple values
	// for the same message-header is supported.
	HTTPHeaders map[string]string `yaml:"httpHeaders,omitempty"`

	// If useHTTPS is true and this option is also true, the exporter's TLS
	// cert will not be verified.
	SkipVerify bool `yaml:"skipVerify"`

	// If useHTTPS is true and skipVerify is true, the sniServerName is used
	// to verify the hostname on the returned certificates.
	// It is also included in the client's handshake to support virtual hosting
	// unless it is an IP address.
	SNIServerName string `yaml:"sniServerName"`

	// Path to the CA cert that has signed the TLS cert, unnecessary
	// if `skipVerify` is set to false.
	CACertPath string `yaml:"caCertPath"`
	// Path to the client TLS cert to use for TLS required connections
	ClientCertPath string `yaml:"clientCertPath"`
	// Path to the client TLS key to use for TLS required connections
	ClientKeyPath string `yaml:"clientKeyPath"`
}

// Scheme returns https if enabled, otherwise http
func (h *HTTPConfig) Scheme() string {
	if h.UseHTTPS {
		return "https"
	}
	return "http"
}

func (h *HTTPConfig) DefaultPort() uint16 {
	if h.UseHTTPS {
		return 443
	}
	return 80
}

// Build returns a configured http.Client
func (h *HTTPConfig) Build() (*http.Client, error) {
	roundTripper, err := func() (http.RoundTripper, error) {
		transport := http.DefaultTransport.(*http.Transport).Clone()

		if h.UseHTTPS {
			transport.TLSClientConfig = &tls.Config{
				InsecureSkipVerify: h.SkipVerify,
				ServerName:         h.SNIServerName,
			}
			if _, err := auth.TLSConfig(transport.TLSClientConfig, h.CACertPath, h.ClientCertPath, h.ClientKeyPath); err != nil {
				return nil, err
			}
		}

		return transport, nil
	}()

	if err != nil {
		return nil, err
	}

	if h.Username != "" {
		roundTripper = &auth.TransportWithBasicAuth{
			RoundTripper: roundTripper,
			Username:     h.Username,
			Password:     h.Password,
		}
	}

	if h.HTTPHeaders == nil {
		h.HTTPHeaders = map[string]string{}
	}

	roundTripper = &addHeader{
		rt:      roundTripper,
		headers: h.HTTPHeaders,
	}

	return &http.Client{
		Timeout:   h.HTTPTimeout.AsDuration(),
		Transport: roundTripper,
	}, nil
}

type addHeader struct {
	headers map[string]string
	rt      http.RoundTripper
}

func (h *addHeader) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range h.headers {
		req.Header.Add(k, v)
	}

	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "SignalFx Smart Agent/"+constants.Version)
	}

	return h.rt.RoundTrip(req)
}
