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

// multiPagePerfFetcher allows callers to split up requests for performance data
// for all inventory objects into batches. This is because for large enough
// vSphere deployments, a query for performance data (queryPerf) will fail if
// performance data is requested for all of the inventory objects in one call.
type multiPagePerfFetcher struct {
	gateway  IGateway
	pageSize int
	log      log.FieldLogger
}

func (f *multiPagePerfFetcher) invIterator(inv []*model.InventoryObject, maxSample int32) *invIterator {
	numObjs := len(inv)
	return &invIterator{
		inv:        inv,
		maxSample:  maxSample,
		pageSize:   f.pageSize,
		numInvObjs: numObjs,
		numPages:   f.numPages(numObjs),
		gateway:    f.gateway,
	}
}

func (f *multiPagePerfFetcher) numPages(numObjs int) int {
	return (numObjs + f.pageSize - 1) / f.pageSize
}
