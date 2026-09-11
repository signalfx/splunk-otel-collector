// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package transform

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"testing"
	"time"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/entry"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/testutil"
)

func newTestParser(t *testing.T, regex string, cacheSize uint16) *Parser {
	cfg := NewConfig("test", conf.Transform{Name: "test"})
	cfg.Regex = regex
	if cacheSize > 0 {
		cfg.Cache.Size = cacheSize
	}
	set := componenttest.NewNopTelemetrySettings()
	op, err := cfg.Build(set)
	require.NoError(t, err)
	return op.(*Parser)
}

func TestParserBuildFailure(t *testing.T) {
	cfg := NewConfig("test", conf.Transform{Name: "test"})
	cfg.OnError = "invalid_on_error"
	set := componenttest.NewNopTelemetrySettings()
	_, err := cfg.Build(set)
	require.ErrorContains(t, err, "invalid `on_error` field")
}

func TestParserByteFailure(t *testing.T) {
	parser := newTestParser(t, "^(?P<key>test)", 0)
	_, err := parser.parse([]byte("invalid"))
	require.ErrorContains(t, err, "type '[]uint8' cannot be parsed as regex")
}

func TestParserStringFailure(t *testing.T) {
	parser := newTestParser(t, "^(?P<key>test)", 0)
	_, err := parser.parse("invalid")
	require.ErrorContains(t, err, "regex pattern does not match")
}

func TestParserInvalidType(t *testing.T) {
	parser := newTestParser(t, "^(?P<key>test)", 0)
	_, err := parser.parse([]int{})
	require.ErrorContains(t, err, "type '[]int' cannot be parsed as regex")
}

func TestParserCache(t *testing.T) {
	parser := newTestParser(t, "^(?P<key>cache)", 200)
	defer func() {
		require.NoError(t, parser.Stop())
	}()
	_, err := parser.parse([]int{})
	require.ErrorContains(t, err, "type '[]int' cannot be parsed as regex")
	require.NotNil(t, parser.cache, "expected cache to be configured")
	require.Equal(t, uint16(200), parser.cache.maxSize())
}

func TestParserRegex(t *testing.T) {
	cases := []struct {
		configure func(*Config)
		input     *entry.Entry
		expected  *entry.Entry
		name      string
	}{
		{
			name: "RootString",
			configure: func(p *Config) {
				p.Regex = "a=(?P<a>.*)"
			},
			input: &entry.Entry{
				Body: "a=b",
			},
			expected: &entry.Entry{
				Body: "a=b",
				Attributes: map[string]any{
					"a": "b",
				},
			},
		},
		{
			name: "MemoryCache",
			configure: func(p *Config) {
				p.Regex = "a=(?P<a>.*)"
				p.Cache.Size = 100
			},
			input: &entry.Entry{
				Body: "a=b",
			},
			expected: &entry.Entry{
				Body: "a=b",
				Attributes: map[string]any{
					"a": "b",
				},
			},
		},
		{
			name: "K8sFileCache",
			configure: func(p *Config) {
				p.Regex = `^(?P<pod_name>[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*)_(?P<namespace>[^_]+)_(?P<container_name>.+)-(?P<container_id>[a-z0-9]{64})\.log$`
				p.Cache.Size = 100
			},
			input: &entry.Entry{
				Body: "coredns-5644d7b6d9-mzngq_kube-system_coredns-901f7510281180a402936c92f5bc0f3557f5a21ccb5a4591c5bf98f3ddbffdd6.log",
			},
			expected: &entry.Entry{
				Body: "coredns-5644d7b6d9-mzngq_kube-system_coredns-901f7510281180a402936c92f5bc0f3557f5a21ccb5a4591c5bf98f3ddbffdd6.log",
				Attributes: map[string]any{
					"container_id":   "901f7510281180a402936c92f5bc0f3557f5a21ccb5a4591c5bf98f3ddbffdd6",
					"container_name": "coredns",
					"namespace":      "kube-system",
					"pod_name":       "coredns-5644d7b6d9-mzngq",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := NewConfig("test", conf.Transform{Name: "test"})
			cfg.OutputIDs = []string{"fake"}
			tc.configure(cfg)

			set := componenttest.NewNopTelemetrySettings()
			op, err := cfg.Build(set)
			require.NoError(t, err)

			defer func() {
				require.NoError(t, op.Stop())
			}()

			fake := testutil.NewFakeOutput(t)
			require.NoError(t, op.SetOutputs([]operator.Operator{fake}))

			ots := time.Now()
			tc.input.ObservedTimestamp = ots
			tc.expected.ObservedTimestamp = ots

			err = op.Process(t.Context(), tc.input)
			require.NoError(t, err)

			fake.ExpectEntry(t, tc.expected)
		})
	}
}

func TestBuildParserRegex(t *testing.T) {
	newBasicParser := func() *Config {
		cfg := NewConfig("test", conf.Transform{Name: "test"})
		cfg.OutputIDs = []string{"test"}
		cfg.Regex = "(?P<all>.*)"
		return cfg
	}

	t.Run("BasicConfig", func(t *testing.T) {
		c := newBasicParser()
		set := componenttest.NewNopTelemetrySettings()
		_, err := c.Build(set)
		require.NoError(t, err)
	})

	t.Run("MissingRegexField", func(t *testing.T) {
		c := newBasicParser()
		c.Regex = ""
		set := componenttest.NewNopTelemetrySettings()
		_, err := c.Build(set)
		require.Error(t, err)
	})

	t.Run("InvalidRegexField", func(t *testing.T) {
		c := newBasicParser()
		c.Regex = "())()"
		set := componenttest.NewNopTelemetrySettings()
		_, err := c.Build(set)
		require.Error(t, err)
	})

	t.Run("NoNamedGroups", func(t *testing.T) {
		c := newBasicParser()
		c.Regex = ".*"
		set := componenttest.NewNopTelemetrySettings()
		_, err := c.Build(set)
		require.ErrorContains(t, err, "no named capture groups")
	})

	t.Run("NoNamedGroups", func(t *testing.T) {
		c := newBasicParser()
		c.Regex = "(.*)"
		set := componenttest.NewNopTelemetrySettings()
		_, err := c.Build(set)
		require.ErrorContains(t, err, "no named capture groups")
	})
}

// return 100 unique file names, example:
// dafplsjfbcxoeff-5644d7b6d9-mzngq_kube-system_coredns-901f7510281180a402936c92f5bc0f3557f5a21ccb5a4591c5bf98f3ddbffdd6.log
// rswxpldnjobcsnv-5644d7b6d9-mzngq_kube-system_coredns-901f7510281180a402936c92f5bc0f3557f5a21ccb5a4591c5bf98f3ddbffdd6.log
// lgtemapezqleqyh-5644d7b6d9-mzngq_kube-system_coredns-901f7510281180a402936c92f5bc0f3557f5a21ccb5a4591c5bf98f3ddbffdd6.log
func benchParseInput() (patterns []string) {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz"
	for i := 1; i <= 100; i++ {
		b := make([]byte, 15)
		for i := range b {
			//nolint:gosec // G404: benchParseInput is a test benchmark helper; weak RNG is acceptable
			b[i] = letterBytes[rand.IntN(len(letterBytes))]
		}
		randomStr := string(b)
		p := fmt.Sprintf("%s-5644d7b6d9-mzngq_kube-system_coredns-901f7510281180a402936c92f5bc0f3557f5a21ccb5a4591c5bf98f3ddbffdd6.log", randomStr)
		patterns = append(patterns, p)
	}
	return patterns
}

// Regex used to parse a kubernetes container log file name, which contains the
// pod name, namespace, container name, container.
const benchParsePattern = `^(?P<pod_name>[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*)_(?P<namespace>[^_]+)_(?P<container_name>.+)-(?P<container_id>[a-z0-9]{64})\.log$`

func newTestBenchParser(b *testing.B, cacheSize uint16) *Parser {
	cfg := NewConfig("test", conf.Transform{Name: "test"})
	cfg.Regex = benchParsePattern
	cfg.Cache.Size = cacheSize

	set := componenttest.NewNopTelemetrySettings()
	op, err := cfg.Build(set)
	require.NoError(b, err)
	return op.(*Parser)
}

func benchmarkParseThreaded(b *testing.B, parser *Parser, input []string) {
	wg := sync.WaitGroup{}

	for _, i := range input {
		wg.Add(1)

		go func(i string) {
			if _, err := parser.match(i); err != nil {
				b.Error(err)
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
}

func benchmarkParse(b *testing.B, parser *Parser, input []string) {
	for _, i := range input {
		if _, err := parser.match(i); err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkParse(b *testing.B) {
	benchParsePatterns := benchParseInput()

	for _, c := range []struct {
		benchFunc func(*testing.B, *Parser, []string)
		name      string
		input     []string
		cacheSize uint16
	}{
		{name: "NoCache", cacheSize: 0, input: benchParsePatterns, benchFunc: benchmarkParseThreaded},
		{name: "WithMemoryCache", cacheSize: 100, input: benchParsePatterns, benchFunc: benchmarkParseThreaded},
		{name: "WithMemoryCacheFullByOne", cacheSize: 99, input: benchParsePatterns, benchFunc: benchmarkParseThreaded},
		{name: "WithMemoryCacheFullBy10", cacheSize: 90, input: benchParsePatterns, benchFunc: benchmarkParseThreaded},
		{name: "WithMemoryCacheFullBy50", cacheSize: 50, input: benchParsePatterns, benchFunc: benchmarkParseThreaded},
		{name: "WithMemoryCacheFullBy90", cacheSize: 10, input: benchParsePatterns, benchFunc: benchmarkParseThreaded},
		{name: "WithMemoryCacheFullBy99", cacheSize: 1, input: benchParsePatterns, benchFunc: benchmarkParseThreaded},
		{name: "NoCacheOneFile", cacheSize: 0, input: []string{benchParsePatterns[0]}, benchFunc: benchmarkParse},
		{name: "WithMemoryCacheOneFile", cacheSize: 100, input: []string{benchParsePatterns[0]}, benchFunc: benchmarkParse},
	} {
		b.Run(c.name, func(b *testing.B) {
			parser := newTestBenchParser(b, c.cacheSize)
			defer require.NoError(b, parser.Stop())
			for b.Loop() {
				c.benchFunc(b, parser, c.input)
			}
		})
	}
}
