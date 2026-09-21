// Copyright Splunk Inc.
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

//go:build linux

package lifecycle

import (
	"context"
	"fmt"

	"github.com/signalfx/splunk-otel-collector/cmd/otelcollauncher/cli"
)

// NewLinux constructs the lifecycle command family for Linux: verbs proxy to
// systemd when it manages the collector service.
// When systemd does not manage the service, or a systemctl query
// fails outright, these commands report a clear error rather than attempting
// to manage the process directly.
func NewLinux() *Manager {
	return &Manager{
		dispatch: func(v verb, args []string, streams cli.IO) int {
			proxy := newSystemdProxy()
			ctx := context.Background()

			managed, err := proxy.managesUnit(ctx)
			if err != nil {
				fmt.Fprintf(streams.Err, "failed to query systemd for %s: %v\n", serviceUnitName, err)
				return cli.ExitFailure
			}
			if !managed {
				fmt.Fprintf(streams.Err, "%s is not managed by systemd; lifecycle commands require systemd\n", serviceUnitName)
				return cli.ExitUnsupported
			}

			return proxy.dispatch(ctx, v, args, streams)
		},
	}
}
