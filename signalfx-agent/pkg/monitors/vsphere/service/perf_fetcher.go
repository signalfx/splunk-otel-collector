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

package service

import (
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/monitors/vsphere/model"
)

type perfFetcher interface {
	invIterator(inv []*model.InventoryObject, maxSample int32) *invIterator
}

// Creates a perfFetcher implementation, either a singlePage or multiPage,
// depending on pageSize (pageSize==0 turns off pagination).
func newPerfFetcher(gateway IGateway, pageSize int, log log.FieldLogger) perfFetcher {
	if pageSize == 0 {
		return &singlePagePerfFetcher{
			gateway: gateway,
			log:     log,
		}
	}
	return &multiPagePerfFetcher{
		gateway:  gateway,
		pageSize: pageSize,
		log:      log,
	}
}
