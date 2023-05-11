package traefik

import (
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	pe "github.com/signalfx/signalfx-agent/pkg/monitors/prometheusexporter"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &pe.Monitor{} }, &pe.Config{})
}
