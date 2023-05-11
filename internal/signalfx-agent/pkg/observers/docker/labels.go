package docker

import (
	"regexp"
	"strings"

	"github.com/docker/go-connections/nat"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

var labelConfigRegexp = regexp.MustCompile(
	`^agent.signalfx.com\.` +
		`(?P<type>monitorType|config|port)` +
		`\.(?P<port>[\w]+)(?:-(?P<port_name>[\w]+))?` +
		`(?:\.(?P<config_key>\w+))?$`)

// LabelConfig contains type and other configurations of a monitor
type LabelConfig struct {
	MonitorType   string
	Configuration map[string]interface{}
}

// ContPort is a struct that contains Port data and a given name
type ContPort struct {
	nat.Port
	Name string
}

// GetConfigLabels converts a set of docker labels into configs organized based on the port
func GetConfigLabels(labels map[string]string) map[ContPort]*LabelConfig {
	portMap := map[ContPort]*LabelConfig{}

	for k, v := range labels {
		if !strings.HasPrefix(k, "agent.signalfx.com") {
			continue
		}

		groups := utils.RegexpGroupMap(labelConfigRegexp, k)
		if groups == nil {
			logger.Errorf("Docker label has invalid agent namespaced key: %s", k)
			continue
		}

		natPort, err := nat.NewPort(nat.SplitProtoPort(groups["port"]))
		if err != nil {
			logger.WithError(err).Errorf("Docker label port '%s' could not be parsed", groups["port"])
			continue
		}

		portObj := ContPort{
			Port: natPort,
			Name: groups["port_name"],
		}

		if _, ok := portMap[portObj]; !ok {
			portMap[portObj] = &LabelConfig{
				Configuration: map[string]interface{}{},
			}
		}

		if groups["type"] == "monitorType" {
			portMap[portObj].MonitorType = v
		} else {
			portMap[portObj].Configuration[groups["config_key"]] = utils.DecodeValueGenerically(v)
		}
	}

	return portMap
}
