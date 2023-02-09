// Copyright  Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
