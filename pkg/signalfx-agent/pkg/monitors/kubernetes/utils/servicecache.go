package utils

import (
	"errors"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type servicesSet map[types.UID]bool

// ServiceCache is used for holding values we care about from a pod
// for quicker lookup than querying the API for them each time.
type ServiceCache struct {
	namespaceSvcUIDCache map[string]servicesSet
	cachedServices       map[types.UID]*CachedService
}

// NewServiceCache creates a new minimal service cache
func NewServiceCache() *ServiceCache {
	return &ServiceCache{
		namespaceSvcUIDCache: make(map[string]servicesSet),
		cachedServices:       make(map[types.UID]*CachedService),
	}
}

// CachedService is used for holding only the neccesarry fields we need
// for label syncing
type CachedService struct {
	UID       types.UID
	Name      string
	Namespace string
	Selector  map[string]string
}

func (cs *CachedService) AsService() *v1.Service {
	if cs == nil {
		return nil
	}

	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			UID:       cs.UID,
			Name:      cs.Name,
			Namespace: cs.Namespace,
		},
		Spec: v1.ServiceSpec{
			Selector: cs.Selector,
		},
	}
}

func newCachedService(svc *v1.Service) *CachedService {
	return &CachedService{
		UID:       svc.UID,
		Name:      svc.Name,
		Namespace: svc.Namespace,
		Selector:  svc.Spec.Selector,
	}
}

// AddService adds or updates a service in cache
func (sc *ServiceCache) AddService(svc *v1.Service) {
	// check if any services exist in this services namespace yet
	if _, exists := sc.namespaceSvcUIDCache[svc.Namespace]; !exists {
		sc.namespaceSvcUIDCache[svc.Namespace] = make(map[types.UID]bool)
	}
	sc.cachedServices[svc.UID] = newCachedService(svc)
	sc.namespaceSvcUIDCache[svc.Namespace][svc.UID] = true
}

func (sc *ServiceCache) Get(uid types.UID) *v1.Service {
	return sc.cachedServices[uid].AsService()
}

// Delete removes a service from the cache
func (sc *ServiceCache) Delete(svc *v1.Service) error {
	return sc.DeleteByKey(svc.UID)
}

// DeleteByKey removes a service from the cache given a UID
func (sc *ServiceCache) DeleteByKey(svcUID types.UID) error {
	cachedService, exists := sc.cachedServices[svcUID]
	if !exists {
		// This could happen if we receive a k8s event out of order
		// For example, if a service is queued to be deleted as the agent starts up
		// and we attempt to delete it before we see it exists from the list/watch
		return errors.New("service does not exist in internal cache")
	}
	delete(sc.namespaceSvcUIDCache[cachedService.Namespace], svcUID)
	if len(sc.namespaceSvcUIDCache[cachedService.Namespace]) == 0 {
		delete(sc.namespaceSvcUIDCache, cachedService.Namespace)
	}
	delete(sc.cachedServices, svcUID)

	return nil
}

func (sc *ServiceCache) GetForNamespace(namespace string) []*v1.Service {
	var out []*v1.Service
	for uid := range sc.namespaceSvcUIDCache[namespace] {
		out = append(out, sc.Get(uid))
	}
	return out
}
