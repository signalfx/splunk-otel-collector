package utils

import (
	"errors"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type jobSet map[types.UID]bool

// JobCache is used for holding values we care about from a job
// for quicker lookup than querying the API for them each time.
type JobCache struct {
	namespaceJobUIDCache map[string]jobSet
	cachedJobs           map[types.UID]*CachedJob
}

// NewJobCache creates a new minimal job cache
func NewJobCache() *JobCache {
	return &JobCache{
		namespaceJobUIDCache: make(map[string]jobSet),
		cachedJobs:           make(map[types.UID]*CachedJob),
	}
}

// CachedJob is used for holding only the neccesarry fields we need
// for label syncing
type CachedJob struct {
	UID             types.UID
	Name            string
	Namespace       string
	OwnerReferences []metav1.OwnerReference
}

func (cj *CachedJob) AsJob() *batchv1.Job {
	if cj == nil {
		return nil
	}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			UID:             cj.UID,
			Name:            cj.Name,
			Namespace:       cj.Namespace,
			OwnerReferences: cj.OwnerReferences,
		},
	}
}

func newCachedJob(job *batchv1.Job) *CachedJob {
	return &CachedJob{
		UID:             job.UID,
		Name:            job.Name,
		Namespace:       job.Namespace,
		OwnerReferences: job.OwnerReferences,
	}
}

// AddJob adds or updates a job in cache
func (jc *JobCache) Add(job *batchv1.Job) {
	// check if any jobs exist in this job namespace yet
	if _, exists := jc.namespaceJobUIDCache[job.Namespace]; !exists {
		jc.namespaceJobUIDCache[job.Namespace] = make(map[types.UID]bool)
	}
	cachedJob := newCachedJob(job)
	jc.cachedJobs[job.UID] = cachedJob
	jc.namespaceJobUIDCache[job.Namespace][job.UID] = true
}

// DeleteByKey removes a job from the cache given a UID
func (jc *JobCache) DeleteByKey(jobUID types.UID) error {
	cachedJob, exists := jc.cachedJobs[jobUID]
	if !exists {
		// could not exist if k8s events come in out of order
		// possible to occur on start up from a race condition
		return errors.New("job does not exist in internal cache")
	}
	delete(jc.namespaceJobUIDCache[cachedJob.Namespace], jobUID)
	if len(jc.namespaceJobUIDCache[cachedJob.Namespace]) == 0 {
		delete(jc.namespaceJobUIDCache, cachedJob.Namespace)
	}
	delete(jc.cachedJobs, jobUID)
	return nil
}

func (jc *JobCache) Get(uid types.UID) *batchv1.Job {
	return jc.cachedJobs[uid].AsJob()
}
