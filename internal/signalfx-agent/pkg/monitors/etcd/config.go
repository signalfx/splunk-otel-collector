package etcd

import "github.com/signalfx/signalfx-agent/pkg/monitors/prometheusexporter"

func init() {
	prometheusexporter.RegisterMonitor(monitorMetadata)
}
