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

//go:build !windows

package main

import (
	"golang.org/x/sys/unix"

	"github.com/signalfx/splunk-otel-collector/internal/opampsupervisor/launcher"
)

// run replaces the launcher process with the selected child process on POSIX
// systems so systemd tracks the collector or supervisor directly.
func run(args, env []string, paths launcher.Paths) error {
	cmd, err := launcher.PrepareCommand(args, env, paths)
	if err != nil {
		return err
	}
	return unix.Exec(cmd.Path, append([]string{cmd.Path}, cmd.Args...), cmd.Env)
}
