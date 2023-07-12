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
