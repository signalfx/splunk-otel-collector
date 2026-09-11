// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package prop

import (
	"fmt"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/featuregates"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/operator/transform"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/entry"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/transformer/copy"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/transformer/noop"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
)

func CreateOperatorConfigs(pCfg conf.Prop, transforms []conf.Transform) []operator.Config {
	var operators []operator.Config
	start := noop.NewConfigWithID(fmt.Sprintf("%s-start", pCfg.Name))
	switch pCfg.Type() {
	case conf.SourceType:
		start.IfExpr = fmt.Sprintf("attributes['sourcetype'] == %q", pCfg.Name)
	case conf.Default:
		// no condition
	case conf.Source:
		start.IfExpr = fmt.Sprintf("attributes['source'] == %q", pCfg.Name)
	case conf.Host:
		start.IfExpr = fmt.Sprintf("attributes['host'] == %q", pCfg.Name)
	default:
		panic(fmt.Sprintf("unknown prop type: %v", pCfg.Type()))
	}
	operators = append(operators, operator.NewConfig(start))
	var previous *helper.WriterConfig
	previous = &start.WriterConfig

	if featuregates.CookFeatureGate.IsEnabled() {
		// TODO implement a split parser.
		//if !pCfg.ShouldLineMerge {
		//
		//}

		for _, tCfg := range pCfg.Transforms {
			for _, stanza := range tCfg.Stanza {
				for _, tDef := range transforms {
					if tDef.Name == stanza {
						t := transform.NewConfig(pCfg.Name, tDef)
						previous.OutputIDs = []string{t.OperatorID}
						operators = append(operators, operator.NewConfig(t))
						previous = &t.WriterConfig
						break
					}
				}
			}
		}

		for _, fa := range pCfg.FieldAliases {
			copyOp := copy.NewConfigWithID(fmt.Sprintf("%s-copy", fa.Name))
			copyOp.From, _ = entry.NewField(fmt.Sprintf("attributes[%q]", fa.From))
			copyOp.To, _ = entry.NewField(fmt.Sprintf("attributes[%q]", fa.To))

			previous.OutputIDs = []string{copyOp.OperatorID}
			operators = append(operators, operator.NewConfig(copyOp))
			previous = &copyOp.WriterConfig
		}

		if pCfg.SourceType != "" {
			sourceTypeOp := copy.NewConfigWithID(fmt.Sprintf("%s-sourcetype", pCfg.Name))

			previous.OutputIDs = []string{sourceTypeOp.OperatorID}
			operators = append(operators, operator.NewConfig(sourceTypeOp))
			previous = &sourceTypeOp.WriterConfig
		}

		// TODO add code to move source, sourcetype, host to cooked fields.
	}

	endNoop := noop.NewConfigWithID(fmt.Sprintf("%s-end", pCfg.Name))
	previous.OutputIDs = []string{endNoop.OperatorID}
	endNoop.OutputIDs = []string{"end"}
	operators = append(operators, operator.NewConfig(endNoop))

	return operators
}
