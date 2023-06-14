package metrics

import (
	"sync"

	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/autoscaling/v2beta1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	k8sutil "github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/utils"
	atypes "github.com/signalfx/signalfx-agent/pkg/monitors/types"
)

type DimensionHandler struct {
	sync.Mutex

	uidKindCache      map[types.UID]string
	sendDimensionFunc func(*atypes.Dimension)

	podCache                *k8sutil.PodCache
	serviceCache            *k8sutil.ServiceCache
	replicaSetCache         *k8sutil.ReplicaSetCache
	jobCache                *k8sutil.JobCache
	updatesForNodeDimension bool
	logger                  log.FieldLogger
}

// NewDimensionHandler creates a handler for dimension updates
func NewDimensionHandler(sendDimensionFunc func(*atypes.Dimension), updatesForNodeDimension bool, logger log.FieldLogger) *DimensionHandler {
	return &DimensionHandler{
		uidKindCache:            make(map[types.UID]string),
		sendDimensionFunc:       sendDimensionFunc,
		podCache:                k8sutil.NewPodCache(),
		serviceCache:            k8sutil.NewServiceCache(),
		replicaSetCache:         k8sutil.NewReplicaSetCache(),
		jobCache:                k8sutil.NewJobCache(),
		updatesForNodeDimension: updatesForNodeDimension,
		logger:                  logger,
	}
}

func (dh *DimensionHandler) HandleAdd(newObj runtime.Object) interface{} {
	var kind string

	switch o := newObj.(type) {
	case *v1.Pod:
		dh.sendDimensionFunc(dimensionForPod(o))
		for _, dim := range dimensionsForPodContainers(o) {
			dh.sendDimensionFunc(dim)
		}
		dh.handleAddPod(o)
		kind = "Pod"
	case *appsv1.DaemonSet:
		dh.sendDimensionFunc(dimensionForDaemonSet(o))
		kind = "DaemonSet"
	case *appsv1.Deployment:
		dh.sendDimensionFunc(dimensionForDeployment(o))
		kind = "Deployment"
	case *appsv1.ReplicaSet:
		dh.handleAddReplicaSet(o)
		dh.sendDimensionFunc(dimensionForReplicaSet(o))
		kind = "ReplicaSet"
	case *v1.ReplicationController:
		dh.sendDimensionFunc(dimensionForReplicationController(o))
		kind = "ReplicationController"
	case *v1.Node:
		for _, dim := range dimensionsForNode(o, dh.updatesForNodeDimension) {
			dh.sendDimensionFunc(dim)
		}
		kind = "Node"
	case *v1.Service:
		dh.handleAddService(o)
		kind = "Service"
	case *appsv1.StatefulSet:
		dh.sendDimensionFunc(dimensionForStatefulSet(o))
		kind = "StatefulSet"
	case *batchv1.Job:
		dh.sendDimensionFunc(dimensionForJob(o))
		dh.handleAddJob(o)
		kind = "Job"
	case *batchv1beta1.CronJob:
		dh.sendDimensionFunc(dimensionForCronJob(o))
		kind = "CronJob"
	case *v2beta1.HorizontalPodAutoscaler:
		dh.sendDimensionFunc(dimensionForHpa(o))
		kind = "HorizontalPodAutoscaler"
	default:
		return nil
	}

	key, err := keyForObject(newObj)
	if err != nil {
		dh.logger.WithFields(log.Fields{
			"error": err,
			"obj":   spew.Sdump(newObj),
		}).Error("Could not get cache key")
		return nil
	}
	if kind != "" {
		dh.uidKindCache[key] = kind
	}

	return key
}

// HandleDelete accepts an object that has been deleted and removes the
// associated datapoints/props from the cache.  MUST HOLD LOCK!!
func (dh *DimensionHandler) HandleDelete(oldObj runtime.Object) interface{} {
	key, err := keyForObject(oldObj)
	if err != nil {
		dh.logger.WithFields(log.Fields{
			"error": err,
			"obj":   spew.Sdump(oldObj),
		}).Error("Could not get cache key")
		return nil
	}
	dh.DeleteByKey(key)
	return key
}

// DeleteByKey delete a cache entry by key.  The supplied interface MUST be the
// same type returned by Handle[Add|Delete].  MUST HOLD LOCK!
func (dh *DimensionHandler) DeleteByKey(key interface{}) {
	cacheKey := key.(types.UID)
	var err error
	switch dh.uidKindCache[cacheKey] {
	case "Pod":
		err = dh.podCache.DeleteByKey(cacheKey)
	case "Service":
		err = dh.handleDeleteService(cacheKey)
	case "ReplicaSet":
		err = dh.replicaSetCache.DeleteByKey(cacheKey)
	case "Job":
		err = dh.jobCache.DeleteByKey(cacheKey)
	}
	if err != nil {
		dh.logger.WithFields(log.Fields{
			"error": err,
			"UID":   cacheKey,
		}).Error("Could not delete key from internal resource cache")
	}
	delete(dh.uidKindCache, cacheKey)
}

func (dh *DimensionHandler) handleAddPod(pod *v1.Pod) {
	dh.podCache.AddPod(pod)

	services := dh.serviceCache.GetForNamespace(pod.Namespace)
	var podServiceNames []string
	for _, ser := range services {
		if k8sutil.SelectorMatchesPod(ser.Spec.Selector, pod) {
			podServiceNames = append(podServiceNames, ser.Name)
		}
	}
	if len(podServiceNames) != 0 {
		if dim := dimensionForPodServices(pod, podServiceNames, true); dim != nil {
			dh.sendDimensionFunc(dim)
		}
	}

	// Check if pod owner of type ReplicaSet is cached. Check owners reference on ReplicaSet to see if it was
	// created by a Deployment. Sync properties accordingly
	rsRef := k8sutil.FindOwnerWithKind(pod.OwnerReferences, "ReplicaSet")
	if rsRef != nil {
		if replicaSet := dh.replicaSetCache.Get(rsRef.UID); replicaSet != nil {
			if deployRef := k8sutil.FindOwnerWithKind(replicaSet.OwnerReferences, "Deployment"); deployRef != nil {
				dh.sendDimensionFunc(dimensionForPodDeployment(pod, deployRef.Name, deployRef.UID))
			} else {
				dh.sendDimensionFunc(dimensionForPodWorkload(pod, replicaSet.Name, "ReplicaSet"))
			}
		}
	}

	// Check if pod owner of type Job is cached. Check owners reference on Job to see if it was
	// created by a CronJob. Sync properties accordingly
	jobRef := k8sutil.FindOwnerWithKind(pod.OwnerReferences, "Job")
	if jobRef != nil {
		if job := dh.jobCache.Get(jobRef.UID); job != nil {
			if cronJobRef := k8sutil.FindOwnerWithKind(job.OwnerReferences, "CronJob"); cronJobRef != nil {
				dh.sendDimensionFunc(dimensionForPodCronJob(pod, cronJobRef.Name, cronJobRef.UID))
			} else {
				dh.sendDimensionFunc(dimensionForPodWorkload(pod, job.Name, "Job"))
			}
		}
	}
}

// handleAddService adds a service to internal cache and, if needed,
// will check cached pods if the service matches to add service name
// and service UID properties to the pod.
func (dh *DimensionHandler) handleAddService(service *v1.Service) {
	dh.serviceCache.AddService(service)
	for _, pod := range dh.podCache.GetForNamespace(service.Namespace) {
		if k8sutil.SelectorMatchesPod(service.Spec.Selector, pod) {
			dim := dimensionForPodServices(pod, []string{service.Name}, true)
			dh.sendDimensionFunc(dim)
		}
	}
}

// handleDeleteService will remove a service from the internal cache
// and remove the service tags on it's matching pods.
func (dh *DimensionHandler) handleDeleteService(uid types.UID) error {
	service := dh.serviceCache.Get(uid)
	if service == nil {
		return nil
	}

	err := dh.serviceCache.DeleteByKey(uid)
	if err != nil {
		return err
	}
	for _, pod := range dh.podCache.GetForNamespace(service.Namespace) {
		if k8sutil.SelectorMatchesPod(service.Spec.Selector, pod) {
			dim := dimensionForPodServices(pod, []string{service.Name}, false)
			dh.sendDimensionFunc(dim)
		}
	}
	return nil
}

// handleAddReplicaSet adds a replicaset to the internal cache
// and emits dimensions related to replicaSets for pods. ReplicaSets should always be
// created before Pods, but in case we receive the event out of order we can still
// check pods in the namespace to sync their relevant properties.
// A potential optimization would be to not check this, and assume we received events in order.
func (dh *DimensionHandler) handleAddReplicaSet(rs *appsv1.ReplicaSet) {
	dh.replicaSetCache.Add(rs)
	deployRef := k8sutil.FindOwnerWithKind(rs.OwnerReferences, "Deployment")
	for _, pod := range dh.podCache.GetForNamespace(rs.Namespace) {
		if rsRef := k8sutil.FindOwnerWithUID(pod.OwnerReferences, rs.UID); rsRef == nil {
			continue
		}
		if deployRef == nil {
			dh.sendDimensionFunc(dimensionForPodWorkload(pod, rs.Name, "ReplicaSet"))
		} else {
			dh.sendDimensionFunc(dimensionForPodDeployment(pod, deployRef.Name, deployRef.UID))
		}
	}
}

// handleAddJob adds a job to the internal cache and emits the dim
// for the given job. Jobs should always be created before pods, but incase we receive
// the job out of sync, we can still check if the pod came in before the job.
// Potential optimization would be to not check this and always assume they come in order.
func (dh *DimensionHandler) handleAddJob(job *batchv1.Job) {
	dh.jobCache.Add(job)
	cronJobRef := k8sutil.FindOwnerWithKind(job.OwnerReferences, "CronJob")
	for _, pod := range dh.podCache.GetForNamespace(job.Namespace) {
		if jobRef := k8sutil.FindOwnerWithUID(pod.OwnerReferences, job.UID); jobRef == nil {
			continue
		}
		if cronJobRef == nil {
			dh.sendDimensionFunc(dimensionForPodWorkload(pod, job.Name, "Job"))
		} else {
			dh.sendDimensionFunc(dimensionForPodCronJob(pod, cronJobRef.Name, cronJobRef.UID))
		}
	}
}
