package kubernetes

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"

	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// AuthType describes the type of authentication to use for the K8s API
type AuthType string

const (
	// AuthTypeNone means no auth is required
	AuthTypeNone AuthType = "none"
	// AuthTypeTLS means to use client TLS certs
	AuthTypeTLS AuthType = "tls"
	// AuthTypeServiceAccount means to use the built-in service account that
	// K8s automatically provisions for each pod.
	AuthTypeServiceAccount AuthType = "serviceAccount"
	// AuthTypeKubeConfig uses local credentials like those used by kubectl.
	AuthTypeKubeConfig AuthType = "kubeConfig"
)

var authTypes = map[AuthType]bool{
	AuthTypeNone:           true,
	AuthTypeTLS:            true,
	AuthTypeServiceAccount: true,
	AuthTypeKubeConfig:     true,
}

// APIConfig contains options relevant to connecting to the K8s API
type APIConfig struct {
	AuthType       AuthType `yaml:"authType" default:"serviceAccount"`
	ClientCertPath string   `yaml:"clientCertPath"`
	ClientKeyPath  string   `yaml:"clientKeyPath"`
	CACertPath     string   `yaml:"caCertPath"`
	SkipVerify     bool     `yaml:"skipVerify" default:"false"`
}

// Validate validates the K8s API config
func (c *APIConfig) Validate() error {
	if !authTypes[c.AuthType] {
		return errors.New("Invalid authType for kubernetes: " + string(c.AuthType))
	}

	if c.AuthType == AuthTypeTLS && (c.ClientCertPath == "" || c.ClientKeyPath == "") {
		return errors.New("for TLS auth, you must set both the kubernetesAPI.clientCertPath " +
			"and kubernetesAPI.clientKeyPath config values")
	}
	return nil
}

// CreateRestConfig creates an Kubernetes API config from user configuration.
func CreateRestConfig(apiConf *APIConfig) (*rest.Config, error) {
	var authConf *rest.Config
	var err error

	if apiConf == nil {
		authConf, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {

		authType := apiConf.AuthType

		var k8sHost string
		if authType != AuthTypeKubeConfig {
			host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
			if len(host) == 0 || len(port) == 0 {
				return nil, fmt.Errorf("unable to load k8s config, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined")
			}
			k8sHost = "https://" + net.JoinHostPort(host, port)
		}

		switch authType {
		// Mainly for testing purposes
		case AuthTypeKubeConfig:
			loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
			configOverrides := &clientcmd.ConfigOverrides{}
			authConf, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
				loadingRules, configOverrides).ClientConfig()

			if err != nil {
				return nil, err
			}
		// Mainly for testing purposes
		case AuthTypeNone:
			authConf = &rest.Config{
				Host: k8sHost,
			}
			authConf.Insecure = true
		case AuthTypeTLS:
			authConf = &rest.Config{
				Host: k8sHost,
				TLSClientConfig: rest.TLSClientConfig{
					CertFile: apiConf.ClientCertPath,
					KeyFile:  apiConf.ClientKeyPath,
					CAFile:   apiConf.CACertPath,
				},
			}
		case AuthTypeServiceAccount:
			// This should work for most clusters but other auth types can be added
			authConf, err = rest.InClusterConfig()
			if err != nil {
				return nil, err
			}
		}
	}

	authConf.Wrap(func(rt http.RoundTripper) http.RoundTripper {
		// Don't use system proxy settings since the API is local to the
		// cluster
		if t, ok := rt.(*http.Transport); ok {
			t2 := t.Clone()
			t2.Proxy = nil
			return t2
		}
		return rt
	})

	if apiConf != nil {
		authConf.Insecure = apiConf.SkipVerify
	}
	return authConf, nil
}

// MakeClient can take configuration if needed for other types of auth
func MakeClient(apiConf *APIConfig) (*k8s.Clientset, error) {
	authConf, err := CreateRestConfig(apiConf)
	if err != nil {
		return nil, err
	}

	client, err := k8s.NewForConfig(authConf)
	if err != nil {
		return nil, err
	}

	return client, nil
}
