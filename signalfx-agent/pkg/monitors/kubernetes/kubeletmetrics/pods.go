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

package kubeletmetrics

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (m *Monitor) getPodsByUID() (map[types.UID]*v1.Pod, error) {
	req, err := m.kubeletClient.NewRequest("GET", "/pods", nil)
	if err != nil {
		return nil, err
	}

	var pods v1.PodList
	err = m.kubeletClient.DoRequestAndSetValue(req, &pods)
	if err != nil {
		return nil, fmt.Errorf("failed to get pods from Kubelet URL %q: %v", req.URL.String(), err)
	}

	byUID := map[types.UID]*v1.Pod{}
	for i := range pods.Items {
		p := &pods.Items[i]
		byUID[p.UID] = p
	}

	return byUID, nil
}

func containerStatusBySpecName(conts []v1.ContainerStatus) map[string]*v1.ContainerStatus {
	byName := map[string]*v1.ContainerStatus{}

	for i := range conts {
		byName[conts[i].Name] = &conts[i]
	}

	return byName
}
