package monitors

import (
	"os"
	"strings"

	"github.com/signalfx/signalfx-agent/pkg/core/services"
)

func setNoProxyEnvvar(value string) {
	os.Setenv("no_proxy", value)
	os.Setenv("NO_PROXY", value)
}

func isProxying() bool {
	return os.Getenv("http_proxy") != "" || os.Getenv("https_proxy") != "" ||
		os.Getenv("HTTP_PROXY") != "" || os.Getenv("HTTPS_PROXY") != ""
}

func getNoProxyEnvvar() string {
	noProxy := os.Getenv("NO_PROXY")
	if noProxy == "" {
		noProxy = os.Getenv("no_proxy")
	}
	return noProxy
}

// sets the service IPs/hostnames in the no_proxy environment variable
func ensureProxyingDisabledForService(service services.Endpoint) {
	host := service.Core().Host
	if isProxying() && len(host) > 0 {
		serviceIP := host
		noProxy := getNoProxyEnvvar()

		for _, existingIP := range strings.Split(noProxy, ",") {
			if existingIP == serviceIP {
				return
			}
		}

		setNoProxyEnvvar(noProxy + "," + serviceIP)
	}
}
