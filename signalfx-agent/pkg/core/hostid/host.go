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

package hostid

import (
	"os"

	fqdn "github.com/Showmax/go-fqdn"
	log "github.com/sirupsen/logrus"
)

func getHostname(useFullyQualifiedHost bool) string {
	var host string
	if useFullyQualifiedHost {
		log.Info("Trying to get fully qualified hostname")
		host = fqdn.Get()
		if host == "unknown" || host == "localhost" {
			log.Info("Error getting fully qualified hostname, using plain hostname")
			host = ""
		}
	}

	if host == "" {
		var err error
		host, err = os.Hostname()
		if err != nil {
			log.Error("Error getting system simple hostname, cannot set hostname")
			return ""
		}
	}

	log.Infof("Using hostname %s", host)
	return host
}
