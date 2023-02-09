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

type singlePagePerfFetcher struct {
	gateway IGateway
	log     log.FieldLogger
}

func (f *singlePagePerfFetcher) invIterator(
	inv []*model.InventoryObject,
	maxSample int32,
) *invIterator {
	numObjs := len(inv)
	return &invIterator{
		inv:        inv,
		maxSample:  maxSample,
		pageSize:   numObjs,
		numInvObjs: numObjs,
		numPages:   1,
		gateway:    f.gateway,
	}
}
