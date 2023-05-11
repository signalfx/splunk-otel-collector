package utils

import (
	"errors"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type replicasetSet map[types.UID]bool

// ReplicaSetCache is used for holding values we care about from a replicaset
// for quicker lookup than querying the API for them each time.
type ReplicaSetCache struct {
	namespaceRsUIDCache map[string]replicasetSet
	cachedReplicaSets   map[types.UID]*CachedReplicaSet
}

// NewReplicaSetCache creates a new replicaset cache
func NewReplicaSetCache() *ReplicaSetCache {
	return &ReplicaSetCache{
		namespaceRsUIDCache: make(map[string]replicasetSet),
		cachedReplicaSets:   make(map[types.UID]*CachedReplicaSet),
	}
}

// CachedReplicaSet is used for holding only the neccesarry fields we need
// for syncing deployment name and UID to pods
type CachedReplicaSet struct {
	UID             types.UID
	Name            string
	Namespace       string
	OwnerReferences []metav1.OwnerReference
}

func (crs *CachedReplicaSet) AsReplicaSet() *appsv1.ReplicaSet {
	if crs == nil {
		return nil
	}

	return &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			UID:             crs.UID,
			Name:            crs.Name,
			Namespace:       crs.Namespace,
			OwnerReferences: crs.OwnerReferences,
		},
	}
}

func newCachedReplicaSet(rs *appsv1.ReplicaSet) *CachedReplicaSet {
	return &CachedReplicaSet{
		UID:             rs.UID,
		Name:            rs.Name,
		Namespace:       rs.Namespace,
		OwnerReferences: rs.OwnerReferences,
	}
}

// AddReplicaSet adds or updates a replicaset in the cache
func (rsc *ReplicaSetCache) Add(rs *appsv1.ReplicaSet) {
	// check if any replicaset exist in this replicaset namespace yet
	if _, exists := rsc.namespaceRsUIDCache[rs.Namespace]; !exists {
		rsc.namespaceRsUIDCache[rs.Namespace] = make(map[types.UID]bool)
	}
	rsc.cachedReplicaSets[rs.UID] = newCachedReplicaSet(rs)
	rsc.namespaceRsUIDCache[rs.Namespace][rs.UID] = true
}

func (rsc *ReplicaSetCache) Get(uid types.UID) *appsv1.ReplicaSet {
	return rsc.cachedReplicaSets[uid].AsReplicaSet()
}

// Delete removes a replicaset from the cache
func (rsc *ReplicaSetCache) Delete(rs *appsv1.ReplicaSet) error {
	return rsc.DeleteByKey(rs.UID)
}

// DeleteByKey removes a replicaset from the cache given a UID
func (rsc *ReplicaSetCache) DeleteByKey(rsUID types.UID) error {
	cachedRs, exists := rsc.cachedReplicaSets[rsUID]
	if !exists {
		// This could happen if we receive a k8s event out of order
		// For example, if a replicaSet is queued to be deleted as the agent starts up
		// and we attempt to delete it before we see it exists from the list/watch
		return errors.New("replicaset does not exist in internal cache")
	}
	delete(rsc.namespaceRsUIDCache[cachedRs.Namespace], rsUID)
	delete(rsc.cachedReplicaSets, rsUID)
	return nil
}

// GetByNamespace returns all cached replica sets within a given namespace.
func (rsc *ReplicaSetCache) GetForNamespace(namespace string) []*appsv1.ReplicaSet {
	var out []*appsv1.ReplicaSet
	for rsUID := range rsc.namespaceRsUIDCache[namespace] {
		out = append(out, rsc.cachedReplicaSets[rsUID].AsReplicaSet())
	}
	return out
}
