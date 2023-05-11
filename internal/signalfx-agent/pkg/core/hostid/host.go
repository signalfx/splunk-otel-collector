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
