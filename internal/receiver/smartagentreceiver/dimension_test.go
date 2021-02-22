// Copyright 2021 Splunk, Inc.
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
	"testing"

	metadata "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/stretchr/testify/assert"
)

func TestDimensionToMetadataUpdate(t *testing.T) {
	dimension := types.Dimension{
		Name:  "my_dimension",
		Value: "my_dimension_value",
		Properties: map[string]string{
			"this_property_should_be_updated": "with_this_property_value",
			"this_property_should_be_removed": "",
		},
		Tags: map[string]bool{
			"this_tag_should_be_added":   true,
			"this_tag_should_be_removed": false,
		},
	}
	metadataUpdate := dimensionToMetadataUpdate(dimension)
	assert.Equal(t, "my_dimension", metadataUpdate.ResourceIDKey)
	assert.Equal(t, metadata.ResourceID("my_dimension_value"), metadataUpdate.ResourceID)

	expectedMetadataToUpdate := map[string]string{
		"this_property_should_be_updated": "with_this_property_value",
	}
	assert.Equal(t, expectedMetadataToUpdate, metadataUpdate.MetadataToUpdate)

	expectedMetadataToAdd := map[string]string{
		"this_tag_should_be_added": "",
	}
	assert.Equal(t, expectedMetadataToAdd, metadataUpdate.MetadataToAdd)

	expectedMetadataToRemove := map[string]string{
		"this_property_should_be_removed": "sf_delete_this_property",
		"this_tag_should_be_removed":      "",
	}
	assert.Equal(t, expectedMetadataToRemove, metadataUpdate.MetadataToRemove)
}

func TestPersistingPropertyAndTagWithSameName(t *testing.T) {
	dimension := types.Dimension{
		Properties: map[string]string{
			"shared_name": "property_value",
		},
		Tags: map[string]bool{
			"shared_name": true,
		},
	}
	metadataUpdate := dimensionToMetadataUpdate(dimension)
	expectedMetadataToUpdate := map[string]string{
		"shared_name": "property_value",
	}
	assert.Equal(t, expectedMetadataToUpdate, metadataUpdate.MetadataToUpdate)

	expectedMetadataToAdd := map[string]string{
		"shared_name": "",
	}
	assert.Equal(t, expectedMetadataToAdd, metadataUpdate.MetadataToAdd)

	assert.Empty(t, metadataUpdate.MetadataToRemove)
}

func TestPersistingPropertyAndRemovedTagWithSameName(t *testing.T) {
	dimension := types.Dimension{
		Properties: map[string]string{
			"shared_name": "property_value",
		},
		Tags: map[string]bool{
			"shared_name": false,
		},
	}
	metadataUpdate := dimensionToMetadataUpdate(dimension)
	expectedMetadataToUpdate := map[string]string{
		"shared_name": "property_value",
	}
	assert.Equal(t, expectedMetadataToUpdate, metadataUpdate.MetadataToUpdate)

	assert.Empty(t, metadataUpdate.MetadataToAdd)

	expectedMetadataToRemove := map[string]string{
		"shared_name": "",
	}
	assert.Equal(t, expectedMetadataToRemove, metadataUpdate.MetadataToRemove)
}

func TestRemovedPropertyAndPersistingTagWithSameName(t *testing.T) {
	dimension := types.Dimension{
		Properties: map[string]string{
			"shared_name": "",
		},
		Tags: map[string]bool{
			"shared_name": true,
		},
	}
	metadataUpdate := dimensionToMetadataUpdate(dimension)
	assert.Empty(t, metadataUpdate.MetadataToUpdate)

	expectedMetadataToAdd := map[string]string{
		"shared_name": "",
	}
	assert.Equal(t, expectedMetadataToAdd, metadataUpdate.MetadataToAdd)

	expectedMetadataToRemove := map[string]string{
		"shared_name": "sf_delete_this_property",
	}
	assert.Equal(t, expectedMetadataToRemove, metadataUpdate.MetadataToRemove)
}

func TestRemovedPropertyAndTagWithSameName(t *testing.T) {
	t.Skipf("Not valid until use case is supported in SFx exporter")
	dimension := types.Dimension{
		Properties: map[string]string{
			"shared_name": "",
		},
		Tags: map[string]bool{
			"shared_name": false,
		},
	}
	metadataUpdate := dimensionToMetadataUpdate(dimension)

	assert.Empty(t, metadataUpdate.MetadataToAdd)
	assert.Empty(t, metadataUpdate.MetadataToUpdate)

	expectedMetadataToRemove := map[string]string{
		"shared_name": "sf_delete_this_property",
	}
	assert.Equal(t, expectedMetadataToRemove, metadataUpdate.MetadataToRemove)
}
