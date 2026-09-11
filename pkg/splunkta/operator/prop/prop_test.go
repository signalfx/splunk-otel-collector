// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package prop

import (
	"testing"

	"go.opentelemetry.io/collector/featuregate"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/featuregates"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/operator/transform"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/transformer/copy"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/transformer/noop"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
)

func TestCreateProps(t *testing.T) {
	require.NoError(t, featuregate.GlobalRegistry().Set(featuregates.CookFeatureGate.ID(), true))
	defer func() {
		require.NoError(t, featuregate.GlobalRegistry().Set(featuregates.CookFeatureGate.ID(), false))
	}()
	ops := CreateOperatorConfigs(conf.Prop{
		Name: "foo",
		Transforms: []conf.PropsTransforms{
			{
				Class:  "calling-foo",
				Stanza: []string{"foo", "bar"},
			},
		},
		FieldAliases: []conf.FieldAlias{
			{
				Name: "foo",
				From: "foo",
				To:   "bar",
			},
			{
				Name: "foobar",
				From: "foo",
				To:   "foobar",
			},
		},
	}, []conf.Transform{
		{
			Name:   "foo",
			Regex:  "foo(.*)",
			Format: "bar$1",
		},
		{
			Name:   "bar",
			Regex:  "bar(.*)",
			Format: "foobar$1",
		},
	})

	require.Len(t, ops, 6)
	require.Equal(t, "foo-start", ops[0].Builder.(*noop.Config).OperatorID)
	require.Equal(t, `transforms-"foo"-"foo"`, ops[1].Builder.(*transform.Config).OperatorID)
	require.Equal(t, "foo-copy", ops[3].Builder.(*copy.Config).OperatorID)
	require.Equal(t, "foobar-copy", ops[4].Builder.(*copy.Config).OperatorID)
	require.Equal(t, `attributes.foobar`, ops[4].Builder.(*copy.Config).To.String())
	require.Equal(t, []string{"foo-end"}, ops[4].Builder.(*copy.Config).OutputIDs)
	require.Equal(t, "foo-end", ops[5].Builder.(*noop.Config).OperatorID)
}

func TestCreatePropsMatchExpressions(t *testing.T) {
	for _, tc := range []struct {
		name     string
		expected string
		prop     conf.Prop
	}{
		{
			name:     "source",
			prop:     conf.Prop{Name: "source::/var/log/app.log"},
			expected: `attributes['source'] == "source::/var/log/app.log"`,
		},
		{
			name:     "sourcetype",
			prop:     conf.Prop{Name: "syslog"},
			expected: `attributes['sourcetype'] == "syslog"`,
		},
		{
			name:     "host",
			prop:     conf.Prop{Name: "host::foo"},
			expected: `attributes['host'] == "host::foo"`,
		},
		{
			name:     "host with case-sensitive option",
			prop:     conf.Prop{Name: "host::(?-i)foo"},
			expected: `attributes['host'] == "host::(?-i)foo"`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ops := CreateOperatorConfigs(tc.prop, nil)
			expr := ops[0].Builder.(*noop.Config).IfExpr
			require.Equal(t, tc.expected, expr)
		})
	}
}
