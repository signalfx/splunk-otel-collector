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
	opampFeatureGateID = "splunk.opamp.enabled"
	opampExtensionKey  = "opamp"
)

var opampFeatureGate = featuregate.GlobalRegistry().MustRegister(
	opampFeatureGateID,
	featuregate.StageAlpha,
	featuregate.WithRegisterDescription("When enabled, the opamp extension is active. "+
		"When disabled (default), the opamp extension is removed from the service.extensions configuration at startup, if it is present."),
	featuregate.WithRegisterFromVersion("v0.151.0"),
)

func RemoveOpAMPIfFeatureGateDisabled(_ context.Context, in *confmap.Conf) error {
	if in == nil {
		return nil
	}

	out := in.ToStringMap()

	service, serviceExtensions, err := getServiceExtensions(out)
	if err != nil {
		return err
	}

	opampInConfig := isOpAMPInServiceExtensions(serviceExtensions)
	gateEnabled := opampFeatureGate.IsEnabled()

	if gateEnabled {
		if !opampInConfig {
			log.Printf("WARNING: Feature gate %q is enabled but the opamp extension is not enabled in the config", opampFeatureGateID)
		}
		return nil
	}

	if !opampInConfig {
		return nil
	}

	log.Printf("INFO: Feature gate %q is disabled: removing opamp extension from config", opampFeatureGateID)
	removeOpAMP(service, serviceExtensions)

	*in = *confmap.NewFromStringMap(out)
	return nil
}

func isOpAMPInServiceExtensions(serviceExtensions []any) bool {
	for _, e := range serviceExtensions {
		if s, ok := e.(string); ok && isOpAMPKey(s) {
			return true
		}
	}
	return false
}

func removeOpAMP(service map[string]any, serviceExtensions []any) {
	filtered := make([]any, 0, len(serviceExtensions))
	for _, e := range serviceExtensions {
		if s, ok := e.(string); ok && isOpAMPKey(s) {
			continue
		}
		filtered = append(filtered, e)
	}
	service["extensions"] = filtered
}

func isOpAMPKey(key string) bool {
	if key == opampExtensionKey {
		return true
	}
	prefix := opampExtensionKey + "/"
	return len(key) > len(prefix) && key[:len(prefix)] == prefix
}
