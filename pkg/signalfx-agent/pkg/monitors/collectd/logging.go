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
