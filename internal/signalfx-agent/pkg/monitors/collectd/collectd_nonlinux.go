//go:build !linux
// +build !linux

package collectd

import (
	"github.com/signalfx/signalfx-agent/pkg/core/config"
)

// ConfigureMainCollectd returns nil on windows because collectd
// does not run on windows
func ConfigureMainCollectd(conf *config.CollectdConfig) error {
	return nil
}
