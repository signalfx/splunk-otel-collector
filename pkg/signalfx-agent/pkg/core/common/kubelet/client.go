package kubelet

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	k8sTransport "k8s.io/client-go/transport"

	"github.com/signalfx/signalfx-agent/pkg/core/common/auth"
)

// AuthType to use when connecting to kubelet
type AuthType string

const (
	// AuthTypeNone means there is no authentication to kubelet
	AuthTypeNone AuthType = "none"
	// AuthTypeTLS indicates that client TLS auth is desired
	AuthTypeTLS AuthType = "tls"
	// AuthTypeServiceAccount indicates that the default service account token should be used
	AuthTypeServiceAccount AuthType = "serviceAccount"
)

// APIConfig contains config specific to the KubeletAPI
type APIConfig struct {
	// URL of the Kubelet instance.  This will default to `http://<current
	// node hostname>:10255` if not provided.
	URL string `yaml:"url"`
	// Can be `none` for no auth, `tls` for TLS client cert auth, or
	// `serviceAccount` to use the pod's default service account token to
	// authenticate.
	AuthType AuthType `yaml:"authType" default:"none"`
	// Whether to skip verification of the Kubelet's TLS cert
	SkipVerify *bool `yaml:"skipVerify" default:"true"`
	// Path to the CA cert that has signed the Kubelet's TLS cert, unnecessary
	// if `skipVerify` is set to false.
	CACertPath string `yaml:"caCertPath"`
	// Path to the client TLS cert to use if `authType` is set to `tls`
	ClientCertPath string `yaml:"clientCertPath"`
	// Path to the client TLS key to use if `authType` is set to `tls`
	ClientKeyPath string `yaml:"clientKeyPath"`
	// Whether to log the raw cadvisor response at the debug level for
	// debugging purposes.
	LogResponses bool `yaml:"logResponses" default:"false"`
}

// Client is a wrapper around http.Client that injects the right auth to every
// request.
type Client struct {
	*http.Client
	config *APIConfig
	logger log.FieldLogger
}

// NewClient creates a new client with the given config
func NewClient(kubeletAPI *APIConfig, logger log.FieldLogger) (*Client, error) {
	if kubeletAPI.URL == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return nil, err
		}
		kubeletAPI.URL = fmt.Sprintf("http://%s:10255", hostname)
	}

	tlsConfig := &tls.Config{}
	if kubeletAPI.SkipVerify != nil {
		tlsConfig.InsecureSkipVerify = *kubeletAPI.SkipVerify
	}

	var transport http.RoundTripper = &(*http.DefaultTransport.(*http.Transport))
	switch kubeletAPI.AuthType {
	case AuthTypeTLS:
		tlsConfig, err := auth.TLSConfig(tlsConfig, kubeletAPI.CACertPath, kubeletAPI.ClientCertPath, kubeletAPI.ClientKeyPath)

		if err != nil {
			return nil, err
		}

		transport.(*http.Transport).TLSClientConfig = tlsConfig
	case AuthTypeServiceAccount:
		certs, err := auth.CertPool()

		if err != nil {
			return nil, err
		}
		tokenPath := "/var/run/secrets/kubernetes.io/serviceaccount/token"
		token, err := ioutil.ReadFile(tokenPath)
		if err != nil {
			return nil, fmt.Errorf("could not read service account token at default location, are "+
				"you sure service account tokens are mounted into your containers by default?: %w", err)
		}

		rootCAFile := "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
		if err := auth.AugmentCertPoolFromCAFile(certs, rootCAFile); err != nil {
			return nil, fmt.Errorf("could not load root CA config from %s: %w", rootCAFile, err)
		}

		tlsConfig.RootCAs = certs
		t := transport.(*http.Transport)
		t.TLSClientConfig = tlsConfig
		transport, err = k8sTransport.NewBearerAuthWithRefreshRoundTripper(string(token), tokenPath, t)
		if err != nil {
			return nil, fmt.Errorf("could not set up refreshable context for kubernetes AuthTypeService: %w", err)
		}
		logger.Debug("Using service account authentication for Kubelet")
	default:
		transport.(*http.Transport).TLSClientConfig = tlsConfig
	}

	return &Client{
		Client: &http.Client{
			Timeout:   10 * time.Second,
			Transport: transport,
		},
		config: kubeletAPI,
		logger: logger,
	}, nil
}

// NewRequest is used to provide a base URL to which paths can be appended.
// Other than the second argument it is identical to the http.NewRequest
// method.
func (kc *Client) NewRequest(method, path string, body io.Reader) (*http.Request, error) {
	baseURL := kc.config.URL
	if !strings.HasSuffix(baseURL, "/") && !strings.HasPrefix(path, "/") {
		baseURL += "/"
	}

	return http.NewRequest(method, baseURL+path, body)
}

// DoRequestAndSetValue does a request to the Kubelet and populates the `value`
// by deserializing the JSON in the response.
func (kc *Client) DoRequestAndSetValue(req *http.Request, value interface{}) error {
	response, err := kc.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read Kubelet response body - %v", err)
	}

	if response.StatusCode == http.StatusNotFound {
		return fmt.Errorf("kubelet request resulted in 404: %s", req.URL.String())
	} else if response.StatusCode != http.StatusOK {
		return fmt.Errorf("kubelet request failed - %q, response: %q", response.Status, string(body))
	}

	if kc.config.LogResponses {
		kc.logger.WithField("url", req.URL.String()).WithField("body", string(body)).Info("Raw response from Kubelet url")
	}

	err = json.Unmarshal(body, value)
	if err != nil {
		return fmt.Errorf("failed to parse Kubelet output. Response: %q. Error: %v", string(body), err)
	}
	return nil
}
