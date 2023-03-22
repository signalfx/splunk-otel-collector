package auth

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"runtime"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func AugmentCertPoolFromCAFile(basePool *x509.CertPool, caCertPath string) error {
	bytes, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		return errors.Wrapf(err, "CA cert path %s could not be read", caCertPath)
	}

	if !basePool.AppendCertsFromPEM(bytes) {
		return errors.Errorf("CA cert file %s is not the right format", caCertPath)
	}

	return nil
}

// Returns a tls.Config that can be used to setup a  tls client
func TLSConfig(tlsConfig *tls.Config, caCertPath string, clientCertPath string, clientKeyPath string) (*tls.Config, error) {
	certs, err := CertPool()

	if err != nil {
		return nil, err
	}

	if caCertPath != "" && certs != nil {
		if err := AugmentCertPoolFromCAFile(certs, caCertPath); err != nil {
			return nil, err
		}
	}

	var clientCerts []tls.Certificate

	if clientCertPath != "" && clientKeyPath != "" {
		cert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
		if err != nil {
			return nil, errors.WithMessage(err,
				fmt.Sprintf("Client cert/key could not be loaded from %s/%s",
					clientKeyPath, clientCertPath))
		}
		clientCerts = append(clientCerts, cert)
		log.Infof("Configured TLS client cert in %s with key %s", clientCertPath, clientKeyPath)
	}

	tlsConfig.Certificates = clientCerts
	tlsConfig.RootCAs = certs

	return tlsConfig, nil
}

// CertPool returns the system cert pool for non-Windows platforms
func CertPool() (*x509.CertPool, error) {
	if runtime.GOOS == "windows" {
		return x509.NewCertPool(), nil
	}

	certs, err := x509.SystemCertPool()
	if err != nil {
		return nil, errors.WithMessage(err, "Could not load system x509 cert pool")
	}
	return certs, nil
}
