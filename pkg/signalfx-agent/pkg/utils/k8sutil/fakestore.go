package k8sutil

import "k8s.io/client-go/tools/cache"

// FixedFakeCustomStore is necessary until we use a client-go version that
// includes https://github.com/kubernetes/kubernetes/pull/62406.
type FixedFakeCustomStore struct {
	cache.FakeCustomStore
}

// Update calls the custom Update function if defined
func (f *FixedFakeCustomStore) Update(obj interface{}) error {
	if f.UpdateFunc != nil {
		return f.UpdateFunc(obj)
	}
	return nil
}
