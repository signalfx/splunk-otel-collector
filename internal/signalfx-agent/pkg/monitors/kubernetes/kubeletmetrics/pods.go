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
