// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package filter provides helpers for translating Splunk whitelist/blacklist
// PCRE regexes into stanza filter operators and filelog include paths.
package filter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/input/file"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/transformer/filter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/split"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/trim"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
)

// NewWhitelistOperator returns a filter operator that drops entries whose
// log.file.path does NOT match the given PCRE regex.
func NewWhitelistOperator(regex string) operator.Config {
	c := filter.NewConfigWithID("whitelist-filter")
	c.Expression = fmt.Sprintf(`!(attributes["log.file.path"] matches %q)`, regex)
	return operator.NewConfig(c)
}

// NewBlacklistOperator returns a filter operator that drops entries whose
// log.file.path matches the given PCRE regex.
func NewBlacklistOperator(regex string) operator.Config {
	c := filter.NewConfigWithID("blacklist-filter")
	c.Expression = fmt.Sprintf(`attributes["log.file.path"] matches %q`, regex)
	return operator.NewConfig(c)
}

// ApplyIncludeExclude sets oc.Include based on the resolved path and whitelist
// param. Whitelist/blacklist are treated as PCRE regexes per Splunk docs and
// are applied as filter operators in BaseConfig — this function only sets the
// filelog include path. It logs the resulting pattern at debug level.
func ApplyIncludeExclude(oc *file.Config, path string, stanza conf.Stanza, receiverName string, logger *zap.Logger) {
	var allowlist string
	switch {
	case strings.ContainsAny(path, "*?["):
		// Path already contains glob metacharacters (e.g. monitor:///home/*/.bash_history);
		// use it directly.
		allowlist = path
	case stanza.Params.Get("whitelist") != nil:
		// whitelist is present (empty or PCRE regex): expand to dir/* so filelog
		// picks up all files; the regex is applied as a filter operator in BaseConfig.
		allowlist = filepath.Join(path, "*")
	default:
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			allowlist = filepath.Join(path, "*")
		} else {
			allowlist = path
		}
	}
	oc.Include = []string{allowlist}
	logger.Debug(
		receiverName+" receiver include pattern",
		zap.String("stanza", stanza.Name),
		zap.String("path", path),
		zap.String("include", allowlist),
	)
}

// ApplyStanzaConfig sets file.Config attributes and defaults that are common
// to both monitor and batch receivers.
func ApplyStanzaConfig(oc *file.Config, stanza conf.Stanza) {
	for _, name := range []string{"host", "index", "sourcetype", "source"} {
		if p := stanza.Params.Get(name); p != nil {
			oc.Attributes[name] = helper.ExprStringConfig(p.Value)
		}
	}
	oc.IncludeFilePath = true
	oc.Encoding = "utf-8"
	oc.StartAt = "beginning"
	oc.SplitConfig = split.Config{
		LineStartPattern: "^",
	}
	oc.TrimConfig = trim.Config{
		PreserveLeading:  true,
		PreserveTrailing: true,
	}
}
