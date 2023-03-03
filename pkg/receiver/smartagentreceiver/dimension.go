// Copyright Splunk, Inc.
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

package smartagentreceiver

import (
	metadata "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
)

// intentionally invalid value as safeguard to prevent backend usage
const deleteThisProperty = "sf_delete_this_property"

// dimensionToMetadataUpdate transforms Smart Agent dimensions to the collector-contrib
// experimental metadata update format to be used by SFx exporter.
// It doesn't comply with agent MergeIntoExisting flag as legacy Get/Put functionality is unnecessary.
func dimensionToMetadataUpdate(dimension types.Dimension) metadata.MetadataUpdate {
	update := metadata.MetadataUpdate{
		ResourceIDKey: dimension.Name,
		ResourceID:    metadata.ResourceID(dimension.Value),
		MetadataDelta: metadata.MetadataDelta{
			MetadataToAdd:    map[string]string{},
			MetadataToRemove: map[string]string{},
			MetadataToUpdate: map[string]string{},
		},
	}
	for property, value := range dimension.Properties {
		if value == "" {
			// SFx Exporter requires value to be not empty to distinguish
			// from tag but actual value is ignored
			update.MetadataToRemove[property] = deleteThisProperty
			continue
		}
		// Doesn't matter if this is first occurrence of property since
		// MetadataToUpdate is generally adequate and reserved for properties
		update.MetadataToUpdate[property] = value
	}

	for tag, keep := range dimension.Tags {
		if !keep {
			update.MetadataToRemove[tag] = ""
			continue
		}
		update.MetadataToAdd[tag] = ""

	}
	return update
}
