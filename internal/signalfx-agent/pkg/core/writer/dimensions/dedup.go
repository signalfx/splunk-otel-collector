package dimensions

import (
	"reflect"

	lru "github.com/hashicorp/golang-lru"
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
)

type deduplicator struct {
	history *lru.Cache
}

func newDeduplicator(size int) *deduplicator {
	history, err := lru.New(size)
	if err != nil {
		panic("could not make dimension cache: " + err.Error())
	}

	return &deduplicator{
		history: history,
	}
}

func (dd *deduplicator) IsDuplicate(dim *types.Dimension) bool {
	cached, ok := dd.history.Get(dim.Key())
	if !ok {
		return false
	}

	cachedDim := cached.(*types.Dimension)

	if cachedDim.MergeIntoExisting != dim.MergeIntoExisting {
		log.Warnf("Dimension %s/%s is updated with both merging and non-merging, which will result in race conditions and inconsistent data.", dim.Name, dim.Value)
		return false
	}

	if !dim.MergeIntoExisting {
		return reflect.DeepEqual(dim, cachedDim)
	}

	// The below checks if any value in the passed in dim is different than
	// what has been last set before.

	for k := range dim.Properties {
		if dim.Properties[k] != cachedDim.Properties[k] {
			return false
		}
	}

	for tag := range dim.Tags {
		cachedDimTag, cachedDimTagOk := cachedDim.Tags[tag]
		if dim.Tags[tag] != cachedDimTag || !cachedDimTagOk {
			return false
		}
	}

	return true
}

// Add the dimension to the deduplicator
func (dd *deduplicator) Add(dim *types.Dimension) {
	cached, ok := dd.history.Get(dim.Key())
	if !ok {
		dd.history.Add(dim.Key(), dim)
		return
	}

	cachedDim := cached.(*types.Dimension)

	if cachedDim.MergeIntoExisting != dim.MergeIntoExisting {
		log.Warnf("Dimension %s/%s is updated with both merging and non-merging, which will result in race conditions and inconsistent data.", dim.Name, dim.Value)
		return
	}

	if !dim.MergeIntoExisting {
		dd.history.Add(dim.Key(), dim)
		return
	}

	// If we are merging dimension props/tags, then just keep all the updates
	// in the cached copy so we will know if an update is going to change
	// anything or not.
	if cachedDim.Properties == nil {
		cachedDim.Properties = map[string]string{}
	}
	for k, v := range dim.Properties {
		// Update the dim in the cache in place
		cachedDim.Properties[k] = v
	}

	if cachedDim.Tags == nil {
		cachedDim.Tags = map[string]bool{}
	}
	for tag, present := range dim.Tags {
		cachedDim.Tags[tag] = present
	}
}
