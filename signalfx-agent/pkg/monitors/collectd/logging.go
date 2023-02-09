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

//go:build linux
// +build linux

package collectd

import (
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/utils"
)

var logRE = regexp.MustCompile(
	`(?s)` + // Allow . to match newlines
		`\[(?P<timestamp>.*?)\] ` +
		`(?:\[(?P<level>\w+?)\] )?` +
		`(?P<message>(?:(?P<plugin>[\w-]+?): )?.*)`)

func logLine(line string, logger log.FieldLogger) {
	groups := utils.RegexpGroupMap(logRE, line)

	var level string
	var message string
	if groups == nil {
		level = "info"
		message = line
	} else {
		if groups["plugin"] != "" {
			logger = logger.WithField("plugin", groups["plugin"])
		}

		level = groups["level"]
		message = strings.TrimPrefix(groups["message"], groups["plugin"]+": ")
	}

	switch level {
	case "debug":
		logger.Debug(message)
	case "info":
		logger.Info(message)
	case "notice":
		logger.Info(message)
	case "warning", "warn":
		logger.Warn(message)
	case "err", "error":
		logger.Error(message)
	default:
		logger.Info(message)
	}
}
