package metrics

import (
	"errors"
	"sync"

	"k8s.io/api/autoscaling/v2beta1"

	"github.com/davecgh/go-spew/spew"
	quota "github.com/openshift/api/quota/v1"
	"github.com/signalfx/golib/v3/datapoint"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

// ContainerID is some type of unique id for containers
type ContainerID string

// DatapointCache holds an up to date copy of datapoints pertaining to the
// cluster.  It is updated whenever the HandleAdd method is called with new
// K8s resources.
type DatapointCache struct {
	sync.Mutex
	dpCache                    map[types.UID][]*datapoint.Datapoint
	nodeConditionTypesToReport []string
	logger                     log.FieldLogger
}

// NewDatapointCache creates a new clean cache
func NewDatapointCache(nodeConditionTypesToReport []string, logger log.FieldLogger) *DatapointCache {
	return &DatapointCache{
		dpCache:                    make(map[types.UID][]*datapoint.Datapoint),
		nodeConditionTypesToReport: nodeConditionTypesToReport,
		logger:                     logger,
	}
}

func keyForObject(obj runtime.Object) (types.UID, error) {
	var key types.UID
	oma, ok := obj.(metav1.ObjectMetaAccessor)
	if !ok || oma.GetObjectMeta() == nil {
		return key, errors.New("kubernetes object is not of the expected form")
	}
	key = oma.GetObjectMeta().GetUID()
	return key, nil
}

// DeleteByKey delete a cache entry by key.  The supplied interface MUST be the
// same type returned by Handle[Add|Delete].  MUST HOLD LOCK!
func (dc *DatapointCache) DeleteByKey(key interface{}) {
	cacheKey := key.(types.UID)
	delete(dc.dpCache, cacheKey)
}

// HandleDelete accepts an object that has been deleted and removes the
// associated datapoints/props from the cache.  MUST HOLD LOCK!!
func (dc *DatapointCache) HandleDelete(oldObj runtime.Object) interface{} {
	key, err := keyForObject(oldObj)
	if err != nil {
		dc.logger.WithFields(log.Fields{
			"error": err,
			"obj":   spew.Sdump(oldObj),
		}).Error("Could not get cache key")
		return nil
	}
	dc.DeleteByKey(key)
	return key
}

// HandleAdd accepts a new (or updated) object and updates the datapoint/prop
// cache as needed.  MUST HOLD LOCK!!
// nolint: funlen
func (dc *DatapointCache) HandleAdd(newObj runtime.Object) interface{} {
	var dps []*datapoint.Datapoint

	switch o := newObj.(type) {
	case *v1.Pod:
		dps = datapointsForPod(o)
	case *v1.Namespace:
		dps = datapointsForNamespace(o)
	case *v1.ReplicationController:
		dps = datapointsForReplicationController(o)
	case *appsv1.DaemonSet:
		dps = datapointsForDaemonSet(o)
	case *appsv1.Deployment:
		dps = datapointsForDeployment(o)
	case *appsv1.ReplicaSet:
		dps = datapointsForReplicaSet(o)
	case *v1.ResourceQuota:
		dps = datapointsForResourceQuota(o)
	case *v1.Node:
		dps = datapointsForNode(o, dc.nodeConditionTypesToReport)
	case *v1.Service:
	case *appsv1.StatefulSet:
		dps = datapointsForStatefulSet(o)
	case *quota.ClusterResourceQuota:
		dps = datapointsForClusterQuotas(o)
	case *batchv1.Job:
		dps = datapointsForJob(o)
	case *batchv1beta1.CronJob:
		dps = datapointsForCronJob(o)
	case *v2beta1.HorizontalPodAutoscaler:
		dps = datapointsForHpa(o, dc.logger)
	default:
		dc.logger.WithFields(log.Fields{
			"obj": spew.Sdump(newObj),
		}).Error("Unknown object type in HandleAdd")
		return nil
	}

	key, err := keyForObject(newObj)
	if err != nil {
		dc.logger.WithFields(log.Fields{
			"error": err,
			"obj":   spew.Sdump(newObj),
		}).Error("Could not get cache key")
		return nil
	}

	if dps != nil {
		dc.dpCache[key] = dps
	}

	return key
}

// AllDatapoints returns all of the cached datapoints.
func (dc *DatapointCache) AllDatapoints() []*datapoint.Datapoint {
	dps := make([]*datapoint.Datapoint, 0)

	dc.Lock()
	defer dc.Unlock()

	for k := range dc.dpCache {
		if dc.dpCache[k] != nil {
			for i := range dc.dpCache[k] {
				// Copy the datapoint since nothing in datapoints is thread
				// safe.
				dp := *dc.dpCache[k][i]
				dps = append(dps, &dp)
			}
		}
	}

	return dps
}
