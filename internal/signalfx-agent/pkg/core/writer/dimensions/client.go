package dimensions

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/apm/requests"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/core/propfilters"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

// DimensionClient sends updates to dimensions to the SignalFx API
type DimensionClient struct {
	sync.RWMutex
	ctx           context.Context
	Token         string
	APIURL        *url.URL
	extraHeaders  map[string]string
	client        *http.Client
	requestSender *requests.ReqSender
	deduplicator  *deduplicator
	sendDelay     time.Duration
	// Set of dims that have been queued up for sending.  Use map to quickly
	// look up in case we need to replace due to flappy prop generation.
	delayedSet map[types.DimensionKey]*types.Dimension
	// Queue of dimensions to update.  The ordering should never change once
	// put in the queue so no need for heap/priority queue.
	delayedQueue      chan *queuedDimension
	PropertyFilterSet *propfilters.FilterSet
	// For easier unit testing
	now        func() time.Time
	logUpdates bool

	DimensionsCurrentlyDelayed int64
	TotalDimensionsDropped     int64
	// The number of dimension updates that happened to the same dimension
	// within sendDelay.
	TotalFlappyUpdates           int64
	TotalClientError4xxResponses int64
	TotalRetriedUpdates          int64
	TotalInvalidDimensions       int64
	TotalDuplicates              int64
}

type queuedDimension struct {
	*types.Dimension
	TimeToSend time.Time
}

// NewDimensionClient returns a new client
func NewDimensionClient(ctx context.Context, conf *config.WriterConfig) (*DimensionClient, error) {
	propFilters, err := conf.PropertyFilters()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:        int(conf.PropertiesMaxRequests),
			MaxIdleConnsPerHost: int(conf.PropertiesMaxRequests),
			IdleConnTimeout:     30 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
	sender := requests.NewReqSender(ctx, client, conf.PropertiesMaxRequests, "dimension")

	return &DimensionClient{
		ctx:               ctx,
		Token:             conf.SignalFxAccessToken,
		APIURL:            conf.ParsedAPIURL(),
		extraHeaders:      conf.ExtraHeaders,
		sendDelay:         time.Duration(conf.PropertiesSendDelaySeconds) * time.Second,
		delayedSet:        make(map[types.DimensionKey]*types.Dimension),
		delayedQueue:      make(chan *queuedDimension, conf.PropertiesMaxBuffered),
		deduplicator:      newDeduplicator(int(conf.PropertiesHistorySize)),
		requestSender:     sender,
		client:            client,
		PropertyFilterSet: propFilters,
		now:               time.Now,
		logUpdates:        conf.LogDimensionUpdates,
	}, nil
}

// Start the client's processing queue
func (dc *DimensionClient) Start() {
	go dc.processQueue()
}

// AcceptDimension to be sent to the API.  This will return fairly quickly and
// won't block.  If the buffer is full, the dim update will be dropped.
func (dc *DimensionClient) AcceptDimension(dim *types.Dimension) error {
	filteredDim := &(*dim)

	filteredDim = dc.PropertyFilterSet.FilterDimension(filteredDim)
	if filteredDim == nil {
		return nil
	}

	dc.Lock()
	defer dc.Unlock()

	if delayedDim := dc.delayedSet[filteredDim.Key()]; delayedDim != nil {
		if !reflect.DeepEqual(delayedDim, filteredDim) {
			dc.TotalFlappyUpdates++

			// Don't further delay it if it gets updated so that we are always
			// guaranteed to update a dimension at least some times, even if it
			// continually gets updated more frequently than `sendDelay` seconds
			// (which should be dealt with somewhere else).

			if filteredDim.MergeIntoExisting != delayedDim.MergeIntoExisting {
				log.Warnf("Dimension %s/%s is updated with both merging and non-merging, which will result in race conditions and inconsistent data.", filteredDim.Name, filteredDim.Value)
			}
			// If the dim is a merge request, then update the existing one in
			// place, otherwise replace it.
			if delayedDim.MergeIntoExisting {
				delayedDim.Properties = utils.MergeStringMaps(delayedDim.Properties, filteredDim.Properties)
				delayedDim.Tags = utils.MergeStringSets(delayedDim.Tags, filteredDim.Tags)
			} else {
				delayedDim.Properties = filteredDim.Properties
				delayedDim.Tags = filteredDim.Tags
			}
		}
	} else {
		atomic.AddInt64(&dc.DimensionsCurrentlyDelayed, int64(1))

		dc.delayedSet[filteredDim.Key()] = filteredDim
		select {
		case dc.delayedQueue <- &queuedDimension{
			Dimension:  filteredDim,
			TimeToSend: dc.now().Add(dc.sendDelay),
		}:
			break
		default:
			dc.TotalDimensionsDropped++
			atomic.AddInt64(&dc.DimensionsCurrentlyDelayed, int64(-1))
			return errors.New("dropped dimension update, propertiesMaxBuffered exceeded")
		}
	}

	return nil
}

func (dc *DimensionClient) processQueue() {
	for {
		select {
		case <-dc.ctx.Done():
			return
		case delayedDim := <-dc.delayedQueue:
			now := dc.now()
			if now.Before(delayedDim.TimeToSend) {
				// dims are always in the channel in order of TimeToSend
				time.Sleep(delayedDim.TimeToSend.Sub(now))
			}

			atomic.AddInt64(&dc.DimensionsCurrentlyDelayed, int64(-1))

			dc.Lock()
			delete(dc.delayedSet, delayedDim.Dimension.Key())
			dc.Unlock()

			if err := dc.setPropertiesOnDimension(delayedDim.Dimension); err != nil {
				log.WithError(utils.SanitizeHTTPError(err)).WithField("dim", delayedDim.Key()).Error("Could not send dimension update")
			}
		}
	}
}

// setPropertiesOnDimension will set custom properties on a specific dimension
// value.  It will wipe out any description on the dimension.
func (dc *DimensionClient) setPropertiesOnDimension(dim *types.Dimension) error {
	var (
		req *http.Request
		err error
	)

	if dim.Name == "" || dim.Value == "" {
		atomic.AddInt64(&dc.TotalInvalidDimensions, int64(1))
		return fmt.Errorf("dimension %v is missing key or value, cannot send", dim)
	}

	if dim.MergeIntoExisting {
		req, err = dc.makePatchRequest(dim.Name, dim.Value, dim.Properties, dim.Tags)
	} else {
		req, err = dc.makeReplaceRequest(dim.Name, dim.Value, dim.Properties, dim.Tags)
	}

	for header, val := range dc.extraHeaders {
		req.Header.Set(header, val)
	}

	if err != nil {
		return err
	}

	req = req.WithContext(
		context.WithValue(req.Context(), requests.RequestFailedCallbackKey, requests.RequestFailedCallback(func(_ []byte, statusCode int, err error) {
			if statusCode >= 400 && statusCode < 500 && statusCode != 404 {
				atomic.AddInt64(&dc.TotalClientError4xxResponses, int64(1))
				log.WithError(utils.SanitizeHTTPError(err)).WithFields(log.Fields{
					"url": req.URL.String(),
					"dim": dim,
				}).Error("Unable to update dimension, not retrying")

				// Don't retry if it is a 4xx error (except 404) since these
				// imply an input/auth error, which is not going to be remedied
				// by retrying.
				// 404 errors are special because they can occur due to races
				// within the dimension patch endpoint.
				return
			}

			log.WithError(utils.SanitizeHTTPError(err)).WithFields(log.Fields{
				"url": req.URL.String(),
				"dim": dim,
			}).Error("Unable to update dimension, retrying")
			atomic.AddInt64(&dc.TotalRetriedUpdates, int64(1))
			// The retry is meant to provide some measure of robustness against
			// temporary API failures.  If the API is down for significant
			// periods of time, dimension updates will probably eventually back
			// up beyond conf.PropertiesMaxBuffered and start dropping.
			if err := dc.AcceptDimension(dim); err != nil {
				log.WithFields(log.Fields{
					"error": utils.SanitizeHTTPError(err),
					"dim":   dim.Key().String(),
				}).Errorf("Failed to retry dimension update")
			}
		})))

	req = req.WithContext(
		context.WithValue(req.Context(), requests.RequestSuccessCallbackKey, requests.RequestSuccessCallback(func([]byte) {
			dc.deduplicator.Add(dim)
			if dc.logUpdates {
				log.WithFields(log.Fields{
					"dim": dim,
				}).Info("Updated dimension")
			}
		})))

	if !dc.deduplicator.IsDuplicate(dim) {
		// This will block if we don't have enough requests
		dc.requestSender.Send(req)
	} else {
		atomic.AddInt64(&dc.TotalDuplicates, int64(1))
	}

	return nil
}

func (dc *DimensionClient) makeDimURL(key, value string) (*url.URL, error) {
	url, err := dc.APIURL.Parse(fmt.Sprintf("/v2/dimension/%s/%s", url.PathEscape(key), url.PathEscape(value)))
	if err != nil {
		return nil, fmt.Errorf("could not construct dimension property PUT URL with %s / %s: %v", key, value, err)
	}

	return url, nil
}

func (dc *DimensionClient) makePatchRequest(key, value string, props map[string]string, tags map[string]bool) (*http.Request, error) {
	tagsToAdd := []string{}
	tagsToRemove := []string{}

	for tag, shouldAdd := range tags {
		if shouldAdd {
			tagsToAdd = append(tagsToAdd, tag)
		} else {
			tagsToRemove = append(tagsToRemove, tag)
		}
	}

	if props == nil {
		props = map[string]string{}
	}

	// The patch endpoint will also accept properties with a null value, in
	// which case they will be deleted, but the agent's Dimension type doesn't
	// support this yet.
	json, err := json.Marshal(map[string]interface{}{
		"customProperties": props,
		"tags":             tagsToAdd,
		"tagsToRemove":     tagsToRemove,
	})
	if err != nil {
		return nil, err
	}

	url, err := dc.makeDimURL(key, value)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"PATCH",
		strings.TrimRight(url.String(), "/")+"/_/sfxagent",
		bytes.NewReader(json))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-SF-TOKEN", dc.Token)

	return req, nil
}

func (dc *DimensionClient) makeReplaceRequest(key, value string, props map[string]string, tags map[string]bool) (*http.Request, error) {
	json, err := json.Marshal(map[string]interface{}{
		"key":              key,
		"value":            value,
		"customProperties": props,
		"tags":             utils.StringSetToSlice(tags),
	})
	if err != nil {
		return nil, err
	}

	url, err := dc.makeDimURL(key, value)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"PUT",
		url.String(),
		bytes.NewReader(json))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-SF-TOKEN", dc.Token)

	return req, nil
}
