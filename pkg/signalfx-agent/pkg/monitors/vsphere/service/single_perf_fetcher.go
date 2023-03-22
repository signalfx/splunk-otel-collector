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
