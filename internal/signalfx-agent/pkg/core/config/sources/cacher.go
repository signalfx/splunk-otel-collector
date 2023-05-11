package sources

import (
	"time"

	"github.com/signalfx/signalfx-agent/pkg/core/config/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	log "github.com/sirupsen/logrus"
)

type configSourceCacher struct {
	source        types.ConfigSource
	cache         map[string]map[string][]byte
	stop          <-chan struct{}
	shouldWatch   bool
	notifications chan<- string
}

func newConfigSourceCacher(source types.ConfigSource, notifications chan<- string, stop <-chan struct{}, shouldWatch bool) *configSourceCacher {
	return &configSourceCacher{
		source:        source,
		stop:          stop,
		shouldWatch:   shouldWatch,
		notifications: notifications,
		cache:         make(map[string]map[string][]byte),
	}
}

// optional controls whether it treats a path not found error as a real error
// that causes watching to never be initiated.
func (c *configSourceCacher) Get(path string, optional bool) (map[string][]byte, error) {
	if v, ok := c.cache[path]; ok {
		return v, nil
	}
	v, version, err := c.source.Get(path)
	if err != nil {
		if _, ok := err.(types.ErrNotFound); !ok || !optional {
			return nil, err
		}
	}

	c.cache[path] = v

	if c.shouldWatch {
		// From now on, subsequent Gets will read from the cache.  It is the
		// responsibility of the watch method to keep the cache up to date.
		go c.watch(path, version)
	}

	return v, nil
}

func (c *configSourceCacher) watch(path string, version uint64) {
	for {
		err := c.source.WaitForChange(path, version, c.stop)
		if utils.IsSignalChanClosed(c.stop) {
			return
		}
		if err != nil {
			// If the file isn't found, just continue
			if _, ok := err.(types.ErrNotFound); ok {
				continue
			}
			log.WithFields(log.Fields{
				"path":   path,
				"source": c.source.Name(),
				"error":  err,
			}).Error("Could not watch path for change")
			time.Sleep(3 * time.Second)
			continue
		}

		values, newVersion, err := c.source.Get(path)
		if err != nil {
			log.WithFields(log.Fields{
				"path":   path,
				"source": c.source.Name(),
				"error":  err,
			}).Error("Could not get path after change")
			version = 0
			time.Sleep(3 * time.Second)
			continue
		}

		version = newVersion
		c.cache[path] = values
		c.notifyChanged(path)
	}
}

func (c *configSourceCacher) notifyChanged(path string) {
	c.notifications <- c.source.Name() + "://" + path
}
