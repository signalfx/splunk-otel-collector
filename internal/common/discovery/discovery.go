// Copyright  Splunk, Inc.
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

package discovery

import (
	"errors"
	"fmt"

	"go.opentelemetry.io/collector/component"
)

const (
	EndpointIDAttr     = "discovery.endpoint.id"
	ObserverIDAttr     = "discovery.observer.id"
	ReceiverConfigAttr = "discovery.receiver.config"
	ReceiverNameAttr   = "discovery.receiver.name"
	ReceiverTypeAttr   = "discovery.receiver.type"
	StatusAttr         = "discovery.status"
	MessageAttr        = "discovery.message"

	OtelEntityTypeAttr        = "otel.entity.type"
	OtelEntityAttributesAttr  = "otel.entity.attributes"
	OtelEntityIDAttr          = "otel.entity.id"
	OtelEntityEventTypeAttr   = "otel.entity.event.type"
	OtelEntityEventTypeState  = "entity_state"
	OtelEntityEventTypeDelete = "entity_delete"
	OtelEntityEventAsLogAttr  = "otel.entity.event_as_log"

	DiscoExtensionsKey = "extensions/splunk.discovery"
	DiscoReceiversKey  = "receivers/splunk.discovery"
)

var NoType = component.MustNewID("SENTINEL_FOR_DISCOVERY_RECEIVER___")

type StatusType string

const (
	Successful StatusType = "successful"
	Partial    StatusType = "partial"
	Failed     StatusType = "failed"
)

var StatusTypes = []StatusType{Successful, Partial, Failed}

var allowedStatuses = func() map[StatusType]struct{} {
	sm := map[StatusType]struct{}{}
	for _, status := range StatusTypes {
		sm[status] = struct{}{}
	}
	return sm
}()

func IsValidStatus(status StatusType) (bool, error) {
	if status == "" {
		return false, errors.New("status cannot be empty")
	}
	if _, ok := allowedStatuses[status]; !ok {
		return false, fmt.Errorf("invalid status %q. must be one of %v", status, StatusTypes)
	}
	return true, nil
}
