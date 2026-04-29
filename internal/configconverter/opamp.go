// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package configconverter

import (
	"context"
	"log"

	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/featuregate"
)

const (
	opampFeatureGateID   = "splunk.opamp.enabled"
	opampSplunkExtension = "opamp/splunk_o11y"
)

var opampFeatureGate = featuregate.GlobalRegistry().MustRegister(
	opampFeatureGateID,
	featuregate.StageAlpha,
	featuregate.WithRegisterDescription("When enabled, the opamp/splunk_o11y extension is active. "+
		"When disabled (default), the opamp/splunk_o11y extension is removed from the service.extensions configuration at startup, if it is present."),
	featuregate.WithRegisterFromVersion("v0.151.0"),
)

func RemoveSplunkOpAMPIfFeatureGateDisabled(_ context.Context, in *confmap.Conf) error {
	if in == nil {
		return nil
	}

	out := in.ToStringMap()

	service, serviceExtensions, err := getServiceExtensions(out)
	if err != nil {
		return err
	}

	opampInConfig := isSplunkOpAMPInServiceExtensions(serviceExtensions)
	gateEnabled := opampFeatureGate.IsEnabled()

	if gateEnabled {
		if !opampInConfig {
			log.Printf("WARNING: Feature gate %q is enabled but %q extension is not enabled in the config", opampFeatureGateID, opampSplunkExtension)
		}
		return nil
	}

	if !opampInConfig {
		return nil
	}

	log.Printf("INFO: Feature gate %q is disabled: removing %q from service.extensions", opampFeatureGateID, opampSplunkExtension)
	removeSplunkOpAMP(service, serviceExtensions)

	*in = *confmap.NewFromStringMap(out)
	return nil
}

func isSplunkOpAMPInServiceExtensions(serviceExtensions []any) bool {
	for _, e := range serviceExtensions {
		if s, ok := e.(string); ok && isSplunkOpAMPExtension(s) {
			return true
		}
	}
	return false
}

func removeSplunkOpAMP(service map[string]any, serviceExtensions []any) {
	filtered := make([]any, 0, len(serviceExtensions))
	for _, e := range serviceExtensions {
		if s, ok := e.(string); ok && isSplunkOpAMPExtension(s) {
			continue
		}
		filtered = append(filtered, e)
	}
	service["extensions"] = filtered
}

func isSplunkOpAMPExtension(key string) bool {
	return key == opampSplunkExtension
}
