// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package filter provides helpers for translating Splunk whitelist/blacklist
// PCRE regexes into stanza filter operators and filelog include paths.
package filter

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp/syntax"
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
// are applied as filter operators in BaseConfig. When either expression is a
// simple alternation of literals, this function also converts it to file globs
// so fileconsumer does not open files that the filters will later discard. The
// regex operators remain the source of truth for entries that are read.
func ApplyIncludeExclude(oc *file.Config, path string, stanza conf.Stanza, receiverName string, logger *zap.Logger) {
	var allowlist []string
	switch {
	case strings.ContainsAny(path, "*?["):
		// Path already contains glob metacharacters (e.g. monitor:///home/*/.bash_history);
		// use it directly.
		allowlist = []string{path}
	case stanza.Params.Get("whitelist") != nil:
		whitelist := stanza.Params.Get("whitelist").Value
		if includes, ok := simpleRegexGlobs(path, whitelist); ok {
			allowlist = includes
		} else {
			// Complex PCRE expressions cannot be represented safely as file globs.
			// Fall back to all files and let the downstream regex operator decide.
			allowlist = []string{filepath.Join(path, "*")}
		}
	default:
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			allowlist = []string{filepath.Join(path, "*")}
		} else {
			allowlist = []string{path}
		}
	}
	oc.Include = allowlist
	if !strings.ContainsAny(path, "*?[") {
		if blacklist := stanza.Params.Get("blacklist"); blacklist != nil {
			if excludes, ok := simpleRegexGlobs(path, blacklist.Value); ok {
				oc.Exclude = excludes
			}
		}
	}
	logger.Debug(
		receiverName+" receiver include pattern",
		zap.String("stanza", stanza.Name),
		zap.String("path", path),
		zap.Strings("include", allowlist),
		zap.Strings("exclude", oc.Exclude),
	)
}

// simpleRegexGlobs converts a safe subset of Splunk path regexes to file globs.
// It intentionally accepts only alternations whose branches are literals with
// an optional end anchor, such as:
//
//	(\.log|log$|messages|secure|auth)
//
// The resulting globs are only a discovery optimization. The original regex
// is still applied after reading, which preserves the TA's filtering behavior.
func simpleRegexGlobs(path, expression string) ([]string, bool) {
	if expression == "" || filepath.Separator != '/' {
		return nil, false
	}

	re, err := syntax.Parse(expression, syntax.Perl)
	if err != nil {
		return nil, false
	}

	branches, ok := simpleLiteralBranches(re)
	if !ok || len(branches) == 0 {
		return nil, false
	}

	includes := make([]string, 0, len(branches))
	seen := make(map[string]struct{}, len(branches))
	for _, branch := range branches {
		// A start-anchored expression or a literal containing a path separator
		// needs full-path regex semantics and cannot be reduced to a basename glob.
		if branch.startAnchored || strings.ContainsRune(branch.literal, filepath.Separator) {
			return nil, false
		}
		// An unanchored literal already present in the monitored directory matches
		// every child path, so narrowing the include list would be incorrect.
		if !branch.endAnchored && strings.Contains(path+string(filepath.Separator), branch.literal) {
			return nil, false
		}

		pattern := "*" + escapeGlobLiteral(branch.literal)
		if !branch.endAnchored {
			pattern += "*"
		}
		include := filepath.Join(path, pattern)
		if _, exists := seen[include]; exists {
			continue
		}
		seen[include] = struct{}{}
		includes = append(includes, include)
	}

	return includes, len(includes) > 0
}

type literalBranch struct {
	literal       string
	startAnchored bool
	endAnchored   bool
}

func simpleLiteralBranches(re *syntax.Regexp) ([]literalBranch, bool) {
	for re.Op == syntax.OpCapture {
		re = re.Sub[0]
	}
	if re.Op != syntax.OpAlternate {
		branch, ok := simpleLiteralBranch(re)
		return []literalBranch{branch}, ok
	}

	branches := make([]literalBranch, 0, len(re.Sub))
	for _, sub := range re.Sub {
		branch, ok := simpleLiteralBranch(sub)
		if !ok {
			return nil, false
		}
		branches = append(branches, branch)
	}
	return branches, true
}

func simpleLiteralBranch(re *syntax.Regexp) (literalBranch, bool) {
	for re.Op == syntax.OpCapture {
		re = re.Sub[0]
	}

	parts := re.Sub
	if re.Op != syntax.OpConcat {
		parts = []*syntax.Regexp{re}
	}

	var branch literalBranch
	for i, part := range parts {
		for part.Op == syntax.OpCapture {
			part = part.Sub[0]
		}
		switch part.Op {
		case syntax.OpBeginText:
			branch.startAnchored = true
		case syntax.OpEndText:
			// An end anchor is only meaningful at the end of the branch. Accepting
			// it earlier would turn an expression such as foo$bar, which cannot
			// match, into the broader glob *foobar.
			if i != len(parts)-1 || branch.endAnchored {
				return literalBranch{}, false
			}
			branch.endAnchored = true
		case syntax.OpLiteral:
			if part.Flags&syntax.FoldCase != 0 {
				return literalBranch{}, false
			}
			branch.literal += string(part.Rune)
		default:
			return literalBranch{}, false
		}
	}
	return branch, branch.literal != ""
}

func escapeGlobLiteral(value string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		`*`, `\*`,
		`?`, `\?`,
		`[`, `\[`,
		`{`, `\{`,
		`}`, `\}`,
	)
	return replacer.Replace(value)
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
	// The zero SplitConfig splits on newlines and drops the terminator, which is
	// what Splunk indexes. A LineStartPattern of "^" ends each token at the next
	// line start instead, keeping the "\n" in _raw and reading linecount as 2.
	oc.SplitConfig = split.Config{}
	oc.TrimConfig = trim.Config{
		PreserveLeading:  true,
		PreserveTrailing: true,
	}
}
