package utils

import (
	"errors"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
)

type podsSet map[types.UID]bool

// PodCache is used for holding values we care about from a pod
// for quicker lookup than querying the API for them each time.
type PodCache struct {
	namespacePodUIDCache map[string]podsSet
	cachedPods           map[types.UID]*CachedPod
}

// CachedPod is used for holding only the necessary
type CachedPod struct {
	UID               types.UID
	LabelSet          labels.Set
	OwnerReferences   []metav1.OwnerReference
	Namespace         string
	Tolerations       []v1.Toleration
	CreationTimestamp metav1.Time
}

func (cp *CachedPod) AsPod() *v1.Pod {
	if cp == nil {
		return nil
	}

	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:               cp.UID,
			Namespace:         cp.Namespace,
			OwnerReferences:   cp.OwnerReferences,
			Labels:            map[string]string(cp.LabelSet),
			CreationTimestamp: cp.CreationTimestamp,
		},
		Spec: v1.PodSpec{
			Tolerations: cp.Tolerations,
		},
	}
}

func newCachedPod(pod *v1.Pod) *CachedPod {
	return &CachedPod{
		UID:               pod.UID,
		LabelSet:          labels.Set(pod.Labels),
		OwnerReferences:   pod.OwnerReferences,
		Namespace:         pod.Namespace,
		Tolerations:       pod.Spec.Tolerations,
		CreationTimestamp: pod.GetCreationTimestamp(),
	}
}

// NewPodCache creates a new minimal pod cache
func NewPodCache() *PodCache {
	return &PodCache{
		namespacePodUIDCache: make(map[string]podsSet),
		cachedPods:           make(map[types.UID]*CachedPod),
	}
}

// AddPod adds or updates a pod in cache
func (pc *PodCache) AddPod(pod *v1.Pod) {
	// check if any pods exist in this pods namespace
	if _, exists := pc.namespacePodUIDCache[pod.Namespace]; !exists {
		pc.namespacePodUIDCache[pod.Namespace] = make(map[types.UID]bool)
	}
	pc.namespacePodUIDCache[pod.Namespace][pod.UID] = true
	pc.cachedPods[pod.UID] = newCachedPod(pod)
}

// DeleteByKey removes a pod from the cache given a UID
func (pc *PodCache) DeleteByKey(key types.UID) error {
	cachedPod, exists := pc.cachedPods[key]
	if !exists {
		// This could happen if we receive a k8s event out of order
		// For example, if a pod deletion event comes in before
		// a pod creation event somehow.
		return errors.New("pod does not exist in internal cache")
	}
	delete(pc.namespacePodUIDCache[cachedPod.Namespace], key)
	if len(pc.namespacePodUIDCache[cachedPod.Namespace]) == 0 {
		delete(pc.namespacePodUIDCache, cachedPod.Namespace)
	}
	delete(pc.cachedPods, key)
	return nil
}

// GetLabels retrieves a pod's cached label set
func (pc *PodCache) GetLabels(key types.UID) labels.Set {
	return pc.cachedPods[key].LabelSet
}

// GetOwnerReferences retrieves a pod's cached owner references
func (pc *PodCache) GetOwnerReferences(key types.UID) []metav1.OwnerReference {
	return pc.cachedPods[key].OwnerReferences
}

// GetPodsInNamespace returns a list of pod UIDs given a namespace
func (pc *PodCache) GetForNamespace(namespace string) []*v1.Pod {
	var pods []*v1.Pod
	for podUID := range pc.namespacePodUIDCache[namespace] {
		pods = append(pods, pc.Get(podUID))
	}
	return pods
}

// GetCachedPod returns a CachedPod object from the cache if it exists
func (pc *PodCache) Get(podUID types.UID) *v1.Pod {
	return pc.cachedPods[podUID].AsPod()
}
