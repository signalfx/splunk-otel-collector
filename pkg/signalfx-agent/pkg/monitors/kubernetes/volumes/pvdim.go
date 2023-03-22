package volumes

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
)

func (m *Monitor) volumeIDDimsForPod(podName, namespace, uid, volName string) (map[string]string, error) {
	// Use pod uid for caching since it is guaranteed to be temporally unique,
	// whereas (podName, namespace) can be reused.
	cacheKey := uid + ":" + volName
	if dims, ok := m.dimCache[cacheKey]; ok {
		return dims, nil
	}

	pod, err := m.k8sClient.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	vol := findVolumeInPod(pod, volName)
	if vol == nil {
		return nil, errors.New("could not find volume in pod spec")
	}

	dims, err := dimsForVolumeSource(vol.VolumeSource, namespace, m.k8sClient)
	if err != nil {
		return nil, err
	}

	m.dimCache[cacheKey] = dims
	return dims, nil
}

func findVolumeInPod(pod *v1.Pod, volName string) *v1.Volume {
	for i := range pod.Spec.Volumes {
		if pod.Spec.Volumes[i].Name == volName {
			return &pod.Spec.Volumes[i]
		}
	}
	return nil
}

func dimsForPersistentVolumeSource(pvs v1.PersistentVolumeSource) map[string]string {
	// IF YOU ADD A NEW PERSISTENT TYPE HERE, ADD IT IN dimsForVolumeSource BELOW TOO!
	switch {
	case pvs.AWSElasticBlockStore != nil:
		return awsElasticBlockStoreDims(*pvs.AWSElasticBlockStore)
	case pvs.GCEPersistentDisk != nil:
		return gcePersistentDiskDims(*pvs.GCEPersistentDisk)
	case pvs.Glusterfs != nil:
		// Special case for pvs.Glusterfs as it is a GlusterfsPersistentVolumeSource instead of GlusterfsVolumeSource
		return glusterfsDims(v1.GlusterfsVolumeSource{
			EndpointsName: pvs.Glusterfs.EndpointsName,
			Path:          pvs.Glusterfs.Path,
			ReadOnly:      pvs.Glusterfs.ReadOnly,
		})
	}
	return nil
}

// Unfortunately client-go uses two separate struct types for directly
// specified volumes and those through PersistentVolumes so we have to
// accommodate both types even though they have largely the same members.
func dimsForVolumeSource(vs v1.VolumeSource, namespace string, client *k8s.Clientset) (map[string]string, error) {
	// IF YOU ADD A NEW PERSISTENT TYPE HERE, ADD IT IN dimsForPersistentVolumeSource
	// ABOVE TOO!
	// PersistentVolumeClaim is unique to VolumeSource, PersistentVolumeSource
	// will not have this.
	switch {
	// Special case for PVC as it will invoke dimsForPersistentVolumeSource lower in the chain:
	case vs.PersistentVolumeClaim != nil:
		return dimsForPersistentVolumeClaim(vs.PersistentVolumeClaim.ClaimName, namespace, client)

	// PERSISTENT TYPES:
	case vs.AWSElasticBlockStore != nil:
		return awsElasticBlockStoreDims(*vs.AWSElasticBlockStore), nil
	case vs.GCEPersistentDisk != nil:
		return gcePersistentDiskDims(*vs.GCEPersistentDisk), nil
	case vs.Glusterfs != nil:
		return glusterfsDims(*vs.Glusterfs), nil

	// NON PERSISTENT TYPES:
	case vs.ConfigMap != nil:
		return map[string]string{"volume_type": "configMap"}, nil
	case vs.DownwardAPI != nil:
		return map[string]string{"volume_type": "downwardAPI"}, nil
	case vs.EmptyDir != nil:
		return map[string]string{"volume_type": "emptyDir"}, nil
	case vs.Secret != nil:
		return map[string]string{"volume_type": "secret"}, nil
	}
	return nil, nil
}

func dimsForPersistentVolumeClaim(claimName, namespace string, client *k8s.Clientset) (map[string]string, error) {
	pvc, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(context.Background(), claimName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	volName := pvc.Spec.VolumeName
	if volName == "" {
		return nil, fmt.Errorf("PersistentVolumeClaim %s does not have a volume name", pvc.Name)
	}

	pv, err := client.CoreV1().PersistentVolumes().Get(context.Background(), volName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return dimsForPersistentVolumeSource(pv.Spec.PersistentVolumeSource), nil
}

func awsElasticBlockStoreDims(vs v1.AWSElasticBlockStoreVolumeSource) map[string]string {
	return map[string]string{
		"volume_type": "awsElasticBlockStore",
		"VolumeId":    vs.VolumeID,
		"fs_type":     vs.FSType,
		"partition":   strconv.Itoa(int(vs.Partition)),
	}
}

func gcePersistentDiskDims(vs v1.GCEPersistentDiskVolumeSource) map[string]string {
	return map[string]string{
		"volume_type": "gcePersistentDisk",
		"pd_name":     vs.PDName,
		"fs_type":     vs.FSType,
		"partition":   strconv.Itoa(int(vs.Partition)),
	}
}

func glusterfsDims(vs v1.GlusterfsVolumeSource) map[string]string {
	return map[string]string{
		"volume_type":    "glusterfs",
		"endpoints_name": vs.EndpointsName,
		"glusterfs_path": vs.Path,
	}
}
