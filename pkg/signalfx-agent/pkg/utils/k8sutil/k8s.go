package k8sutil

import (
	"context"
	"fmt"
	"regexp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/api/core/v1"
	k8s "k8s.io/client-go/kubernetes"
)

var re = regexp.MustCompile(`^[\w_-]+://`)

// EnvValueForContainer returns the value of an env var set on a container
// within a pod.  It only supports simple env values and not value references.
func EnvValueForContainer(pod *v1.Pod, envName, containerName string) (string, error) {
	container := ContainerInPod(pod, containerName)
	for _, env := range container.Env {
		if env.Name == envName {
			if env.ValueFrom != nil {
				return "", fmt.Errorf("container %s env var %s is not a simple value", containerName, envName)
			}
			return env.Value, nil
		}
	}
	return "", fmt.Errorf("container %s does not have env var %s", containerName, envName)
}

// ContainerInPod returns a reference to a container object given its name in
// the given pod.  Returns nil if the container does not exist in the given
// pod.
func ContainerInPod(pod *v1.Pod, containerName string) *v1.Container {
	for i := range pod.Spec.Containers {
		if pod.Spec.Containers[i].Name == containerName {
			return &pod.Spec.Containers[i]
		}
	}
	return nil
}

// FetchSecretValue fetches a specific secret value from the "data" object in a
// secret and returns the decoded value.
func FetchSecretValue(client *k8s.Clientset, secretName, dataKey, namespace string) (string, error) {
	secret, err := client.CoreV1().Secrets(namespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if _, ok := secret.Data[dataKey]; !ok {
		return "", fmt.Errorf("secret value %s not found in secret %s", dataKey, secretName)
	}

	return string(secret.Data[dataKey]), nil
}

// PortByName returns the first port instance with the given name among all of
// a pod's containers.  Returns nil if no port name matches.
func PortByName(pod *v1.Pod, port string) *v1.ContainerPort {
	for _, c := range pod.Spec.Containers {
		for i := range c.Ports {
			p := &c.Ports[i]
			if p.Name == port {
				return p
			}
		}
	}
	return nil
}

// PortByNumber returns the first port instance, regardless of whether it is
// TCP or UDP, with the given port number among all of a pod's containers.
// Returns nil if no port number matches.
func PortByNumber(pod *v1.Pod, port int32) *v1.ContainerPort {
	for _, c := range pod.Spec.Containers {
		for i := range c.Ports {
			p := &c.Ports[i]
			if p.ContainerPort == port {
				return p
			}
		}
	}
	return nil
}

// StripContainerID returns a pure container id without the runtime scheme://
func StripContainerID(id string) string {
	return re.ReplaceAllString(id, "")
}
