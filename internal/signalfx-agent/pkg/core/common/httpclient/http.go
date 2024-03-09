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
	HTTPHeaders    map[string]string `yaml:"httpHeaders,omitempty"`
	Username       string            `yaml:"username"`
	Password       string            `yaml:"password" neverLog:"true"`
	SNIServerName  string            `yaml:"sniServerName"`
	CACertPath     string            `yaml:"caCertPath"`
	ClientCertPath string            `yaml:"clientCertPath"`
	ClientKeyPath  string            `yaml:"clientKeyPath"`
	HTTPTimeout    timeutil.Duration `yaml:"httpTimeout" default:"10s"`
	UseHTTPS       bool              `yaml:"useHTTPS"`
	SkipVerify     bool              `yaml:"skipVerify"`
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
				InsecureSkipVerify: h.SkipVerify, // nolint: gosec
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
