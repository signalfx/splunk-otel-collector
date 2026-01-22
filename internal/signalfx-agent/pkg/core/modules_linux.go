//go:build linux

package core

// Do an import of all of the built-in observers and monitors that are
// not appropriate for windows until we get a proper plugin system

import (
	// Import everything that isn't referenced anywhere else
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/cgroups"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/apache"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/cpu"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/cpufreq"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/custom"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/memcached"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/memory"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/processes"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/protocols"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/python"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/uptime"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/process"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/monitors/varnish"
)
