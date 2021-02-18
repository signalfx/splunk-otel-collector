// Copyright 2021, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package smartagentextension

import (
	"go.opentelemetry.io/collector/config/configmodels"
)

type Config struct {
	configmodels.ExtensionSettings `mapstructure:",squash"`
	BundleDir                      string         `mapstructure:"bundleDir"`
	CollectdConfig                 CollectdConfig `mapstructure:"collectd"`
}

// CollectdConfig fields have a direct mapping to fields on
// collect.CollectdConfig in the SignalFx Agent.
// https://github.com/signalfx/signalfx-agent/blob/v5.7.2/pkg/core/config/config.go#L342
type CollectdConfig struct {
	// How many read intervals before abandoning a metric. Doesn't affect much
	// in normal usage.
	// See [Timeout](https://collectd.org/documentation/manpages/collectd.conf.5.shtml#timeout_iterations).
	Timeout int `mapstructure:"timeout"`
	// Number of threads dedicated to executing read callbacks. See
	// [ReadThreads](https://collectd.org/documentation/manpages/collectd.conf.5.shtml#readthreads_num)
	ReadThreads int `mapstructure:"readThreads"`
	// Number of threads dedicated to writing value lists to write callbacks.
	// This should be much less than readThreads because writing is batched in
	// the write_http plugin that writes back to the agent.
	// See [WriteThreads](https://collectd.org/documentation/manpages/collectd.conf.5.shtml#writethreads_num).
	WriteThreads int `mapstructure:"writeThreads"`
	// The maximum numbers of values in the queue to be written back to the
	// agent from collectd.  Since the values are written to a local socket
	// that the agent exposes, there should be almost no queuing and the
	// default should be more than sufficient. See
	// [WriteQueueLimitHigh](https://collectd.org/documentation/manpages/collectd.conf.5.shtml#writequeuelimithigh_highnum)
	WriteQueueLimitHigh int `mapstructure:"writeQueueLimitHigh"`
	// The lowest number of values in the collectd queue before which metrics
	// begin being randomly dropped.  See
	// [WriteQueueLimitLow](https://collectd.org/documentation/manpages/collectd.conf.5.shtml#writequeuelimitlow_lownum)
	WriteQueueLimitLow int `mapstructure:"writeQueueLimitLow"`
	// Collectd's log level -- info, notice, warning, or err
	LogLevel string `mapstructure:"logLevel"`
	// A default read interval for collectd plugins.  If zero or undefined,
	// will default to the global agent interval.  Some collectd python
	// monitors do not support overridding the interval at the monitor level,
	// but this setting will apply to them.
	IntervalSeconds int `mapstructure:"intervalSeconds"`
	// The local IP address of the server that the agent exposes to which
	// collectd will send metrics.  This defaults to an arbitrary address in
	// the localhost subnet, but can be overridden if needed.
	WriteServerIPAddr string `mapstructure:"writeServerIPAddr"`
	// The port of the agent's collectd metric sink server.  If set to zero
	// (the default) it will allow the OS to assign it a free port.
	WriteServerPort uint16 `mapstructure:"writeServerPort"`
	// This is where the agent will write the collectd collect files that it
	// manages.  If you have secrets in those files, consider setting this to a
	// path on a tmpfs mount.  The files in this directory should be considered
	// transient -- there is no value in editing them by hand.  If you want to
	// add your own collectd collect, see the collectd/custom monitor.
	ConfigDir string `mapstructure:"configDir"`
}
