// Copyright  Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
