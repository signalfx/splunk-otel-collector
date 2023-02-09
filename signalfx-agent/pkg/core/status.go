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

package core

import (
	"fmt"
	"net/http"
	"time"

	"github.com/signalfx/signalfx-agent/pkg/core/common/constants"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/leadership"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

func (a *Agent) diagnosticTextHandler(rw http.ResponseWriter, req *http.Request) {
	section := req.URL.Query().Get("section")
	_, _ = rw.Write([]byte(a.DiagnosticText(section)))
}

// DiagnosticText returns a simple textual output of the agent's status
func (a *Agent) DiagnosticText(section string) string {
	showAll := section == "all"
	var out string
	if section == "" || showAll {
		uptime := time.Since(a.startTime).Round(1 * time.Second).String()
		out +=
			"SignalFx Agent version:           " + constants.Version + "\n" +
				"Agent uptime:                     " + uptime + "\n" +
				a.observers.DiagnosticText() + "\n" +
				a.monitors.SummaryDiagnosticText() + "\n" +
				a.writer.DiagnosticText() + "\n"

		k8sLeader := leadership.CurrentLeader()
		if k8sLeader != "" {
			out += fmt.Sprintf("Kubernetes Leader Node:           %s\n", k8sLeader)
		}

		if section == "" {
			out += "\n" + utils.StripIndent(`
			  Additional status commands:

			  signalfx-agent status config - show resolved config in use by agent
			  signalfx-agent status endpoints - show discovered endpoints
			  signalfx-agent status monitors - show active monitors
			  signalfx-agent status all - show everything
			  `)
		}
	}

	if section == "config" || showAll {
		out += "Agent Configuration:\n" +
			utils.IndentLines(config.ToString(a.lastConfig), 2) + "\n"
	}

	if section == "monitors" || showAll {
		out += a.monitors.DiagnosticText() + "\n"
	}

	if section == "endpoints" || showAll {
		out += a.monitors.EndpointsDiagnosticText()
	}

	return out
}
