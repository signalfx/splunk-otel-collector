// Package kubernetes contains an observer that watches the Kubernetes API for
// pods and nodes that are running on the same node as the agent.  It uses the
// streaming watch API in K8s so that updates are seen immediately without any
// polling interval.
package kubernetes

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/signalfx/signalfx-agent/pkg/core/common/kubernetes"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/core/services"
	"github.com/signalfx/signalfx-agent/pkg/observers"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	"github.com/signalfx/signalfx-agent/pkg/utils/k8sutil"
)

const (
	observerType = "k8s-api"
	nodeEnvVar   = "MY_NODE_NAME"
	runningPhase = "Running"
)

// OBSERVER(k8s-api): Discovers pod endpoints and nodes running in a Kubernetes
// cluster by querying the Kubernetes API server.  This observer by default
// will only discover pod endpoints exposed on the same node that the agent is
// running, so that the monitoring of services does not generate cross-node
// traffic.  To know which node the agent is running on, you should set an
// environment variable called `MY_NODE_NAME` using the downward API
// `spec.nodeName` value in the pod spec. Our provided K8s DaemonSet resource
// does this already and provides an example.
//
// This observer also emits high-level `pod` targets that only contain a `host`
// field but no `port`.  This allows monitoring ports on a pod that are not
// explicitly specified in the pod spec, which would result in no normal
// `hostport` target being emitted for that particular endpoint.
//
// If `discoverAllPods` is set to `true`, then the observer will discover pods on all
// nodes in the cluster (or namespace if specified).  By default, only pods on
// the same node as the agent are discovered.
//
// Note that this observer discovers exposed ports on pod containers, not K8s
// Endpoint resources, so don't let the terminology of agent "endpoints"
// confuse you.

// ENDPOINT_TYPE(ContainerEndpoint): true

// DIMENSION(kubernetes_namespace): The namespace that the discovered service
// endpoint is running in.

// DIMENSION(kubernetes_pod_name): The name of the running pod that is exposing
// the discovered endpoint

// DIMENSION(kubernetes_pod_uid): The UID of the pod that is exposing the
// discovered endpoint

// DIMENSION(container_spec_name): The short name of the container in the pod spec,
// **NOT** the running container's name in the Docker engine

// DIMENSION(kubernetes_node): For Node (`k8s-node`) targets, the name of the Node

// DIMENSION(kubernetes_node_uid): For Node (`k8s-node`) targets, the UID of
// the Node

// ENDPOINT_VAR(kubernetes_annotations|map of strings): The set of annotations on the
// discovered pod or node.

// ENDPOINT_VAR(pod_spec|pod spec): The full pod spec object, as represented by the Go
// K8s client library (client-go): https://godoc.org/k8s.io/api/core/v1#PodSpec.

// ENDPOINT_VAR(pod_metadata|pod metadata): The full pod metadata object, as represented by the Go
// K8s client library (client-go): https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta.

// ENDPOINT_VAR(node_metadata|node_metadata): The metadata about the Node, for `k8s-node`
// targets, with fields in TitleCase.  See [ObjectMeta v1 meta
// reference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#objectmeta-v1-meta).

// ENDPOINT_VAR(node_spec|node_spec): The Node spec object, for `k8s-node` targets.  See
// [the K8s reference on this
// resource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#nodespec-v1-core),
// but keep in the mind that fields will be in TitleCase due to passing through
// Go.

// ENDPOINT_VAR(node_status|node_status): The Node status object, for `k8s-node` targets.
// See [the K8s reference on Node
// Status](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#nodestatus-v1-core)
// but keep in mind that fields will be in TitleCase due to passing through Go.

// ENDPOINT_VAR(node_addresses|map of strings): A map of the different Node addresses
// specified in the Node status object.  The key of the map is the address type
// and the value is the address string. The address types are `Hostname`,
// `ExternalIP`, `InternalIP`, `ExternalDNS`, `InternalDNS`.  Most likely not
// all of these address types will be present for a given Node.

func init() {
	observers.Register(observerType, func(cbs *observers.ServiceCallbacks) interface{} {
		return &Observer{
			serviceCallbacks: cbs,
			endpointsByUID:   make(map[types.UID][]services.Endpoint),
		}
	}, &Config{})
}

// Config for Kubernetes API observer
type Config struct {
	config.ObserverConfig
	// If specified, only pods within the given namespace on the same node as
	// the agent will be discovered. If blank, all pods on the same node as the
	// agent will be discovered.
	Namespace string `yaml:"namespace"`
	// Configuration for the K8s API client
	KubernetesAPI *kubernetes.APIConfig `yaml:"kubernetesAPI" default:"{}"`
	// A list of annotation names that should be used to infer additional ports
	// to be discovered on a particular pod.  The pod's annotation value should
	// be a port number.  This is useful for annotations like `prometheus.io/port: 9230`.
	AdditionalPortAnnotations []string `yaml:"additionalPortAnnotations"`
	// If true, this observer will watch all Kubernetes pods and discover
	// endpoints/services from each of them.  The default behavior (when
	// `false`) is to only watch the pods on the current node that this agent
	// is running on (it knows the current node via the `MY_NODE_NAME` envvar
	// provided by the downward API).
	DiscoverAllPods bool `yaml:"discoverAllPods"`
	// If `true`, the observer will discover nodes as a special type of
	// endpoint.  You can match these endpoints in your discovery rules with
	// the condition `target == "k8s-node"`.
	DiscoverNodes bool `yaml:"discoverNodes"`
}

// Validate the observer-specific config
func (c *Config) Validate() error {
	if err := c.KubernetesAPI.Validate(); err != nil {
		return err
	}

	if os.Getenv(nodeEnvVar) == "" && !c.DiscoverAllPods {
		return fmt.Errorf("kubernetes node name was not provided in the %s envvar", nodeEnvVar)
	}
	return nil
}

// Observer that watches the Kubernetes API for new pods pertaining to this
// node
type Observer struct {
	config           *Config
	clientset        *k8s.Clientset
	thisNode         string
	serviceCallbacks *observers.ServiceCallbacks
	// A cache for endpoints so they don't have to be reconstructed when being
	// removed.
	endpointsByUID map[types.UID][]services.Endpoint
	stopper        chan struct{}
	logger         log.FieldLogger
}

// Configure configures and starts watching for endpoints
func (o *Observer) Configure(config *Config) error {
	o.logger = log.WithFields(log.Fields{"observerType": observerType})

	// There is a bug/limitation in the k8s go client's Controller where
	// goroutines are leaked even when using the stop channel properly.  So we
	// should avoid going through a shutdown/startup cycle here if nothing is
	// different in the config.
	if reflect.DeepEqual(config, o.config) {
		return nil
	}

	o.config = config
	o.thisNode = os.Getenv(nodeEnvVar)

	var err error
	o.clientset, err = kubernetes.MakeClient(config.KubernetesAPI)
	if err != nil {
		return err
	}

	o.stopIfRunning()
	o.watchPods()

	if config.DiscoverNodes {
		o.watchNodes()
	}

	return nil
}

func castPod(obj interface{}) *v1.Pod {
	pod, _ := obj.(*v1.Pod)
	return pod
}

func (o *Observer) watchPods() {
	o.stopper = make(chan struct{})

	podSelector := fields.Everything()
	if !o.config.DiscoverAllPods {
		podSelector = fields.ParseSelectorOrDie("spec.nodeName=" + o.thisNode)
	}

	client := o.clientset.CoreV1().RESTClient()
	watchList := cache.NewListWatchFromClient(client, "pods", o.config.Namespace, podSelector)

	_, controller := cache.NewInformer(
		watchList,
		&v1.Pod{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				o.changeHandler(nil, castPod(obj))
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				o.changeHandler(castPod(oldObj), castPod(newObj))
			},
			DeleteFunc: func(obj interface{}) {
				o.changeHandler(castPod(obj), nil)
			},
		})

	go controller.Run(o.stopper)
}

func castNode(obj interface{}) *v1.Node {
	node, _ := obj.(*v1.Node)
	return node
}

func (o *Observer) watchNodes() {
	o.stopper = make(chan struct{})

	nodeSelector := fields.Everything()

	client := o.clientset.CoreV1().RESTClient()
	watchList := cache.NewListWatchFromClient(client, "nodes", "", nodeSelector)

	_, controller := cache.NewInformer(
		watchList,
		&v1.Node{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				o.changeHandler(nil, castNode(obj))
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				o.changeHandler(castNode(oldObj), castNode(newObj))
			},
			DeleteFunc: func(obj interface{}) {
				o.changeHandler(castNode(obj), nil)
			},
		})

	go controller.Run(o.stopper)
}

func (o *Observer) stopIfRunning() {
	// Stop previous informers
	if o.stopper != nil {
		close(o.stopper)
		o.stopper = nil
	}
}

// Handles notifications of changes to pods from the API server
func (o *Observer) changeHandler(oldObj metav1.ObjectMetaAccessor, newObj metav1.ObjectMetaAccessor) {
	var newEndpoints []services.Endpoint
	var oldEndpoints []services.Endpoint

	if oldObj != nil {
		oldEndpoints = o.endpointsByUID[oldObj.GetObjectMeta().GetUID()]
		delete(o.endpointsByUID, oldObj.GetObjectMeta().GetUID())
	}

	if newObj != nil {
		switch obj := newObj.(type) {
		case *v1.Pod:
			newEndpoints = o.endpointsInPod(obj, o.clientset, utils.StringSliceToMap(o.config.AdditionalPortAnnotations))
		case *v1.Node:
			nodeEndpoint, err := endpointForNode(obj)
			if err != nil {
				o.logger.WithError(err).Warn("Failed to derive endpoint from K8s node")
			} else {
				newEndpoints = append(newEndpoints, nodeEndpoint)
			}
		}

		if len(newEndpoints) > 0 {
			o.endpointsByUID[newObj.GetObjectMeta().GetUID()] = newEndpoints
		}
	}

	// Prevent spurious churn of endpoints if they haven't changed
	if reflect.DeepEqual(newEndpoints, oldEndpoints) {
		return
	}

	// If it is an update, there will be a remove and immediately subsequent
	// add.
	for i := range oldEndpoints {
		o.serviceCallbacks.Removed(oldEndpoints[i])
	}

	for i := range newEndpoints {
		o.serviceCallbacks.Added(newEndpoints[i])
	}
}

func endpointForNode(node *v1.Node) (services.Endpoint, error) {
	id := fmt.Sprintf("node-%s-%s", node.Name, node.UID[:7])
	dims := map[string]string{
		"kubernetes_node":     node.Name,
		"kubernetes_node_uid": string(node.UID),
	}

	endpoint := services.NewEndpointCore(id, node.Name, observerType, dims)

	addrs := map[v1.NodeAddressType]string{}
	for _, addr := range node.Status.Addresses {
		addrs[addr.Type] = addr.Address
	}

	if len(addrs) == 0 {
		return nil, errors.New("failed to determine node IP")
	}
	for _, addrTyp := range []v1.NodeAddressType{
		// These are in priority order
		v1.NodeInternalIP,
		v1.NodeInternalDNS,
		v1.NodeHostName,
		v1.NodeExternalIP,
		v1.NodeExternalDNS,
	} {
		if addrs[addrTyp] != "" {
			endpoint.Host = addrs[addrTyp]
			break
		}
	}

	// Zero out the timestamps in the conditions so that the endpoints don't
	// churn so much.
	for i := range node.Status.Conditions {
		node.Status.Conditions[i].LastHeartbeatTime = metav1.Time{}
		node.Status.Conditions[i].LastTransitionTime = metav1.Time{}
	}
	// Blank out this to avoid unnecessary churn.
	node.ResourceVersion = ""

	endpoint.AddExtraField("kubernetes_annotations", node.Annotations)
	endpoint.AddExtraField("node_metadata", &node.ObjectMeta)
	endpoint.AddExtraField("node_spec", &node.Spec)
	endpoint.AddExtraField("node_status", &node.Status)
	endpoint.AddExtraField("node_addresses", addrs)
	endpoint.Target = services.TargetTypeKubernetesNode

	return endpoint, nil
}

func (o *Observer) endpointsInPod(pod *v1.Pod, client *k8s.Clientset, portAnnotationSet map[string]bool) []services.Endpoint {
	endpoints := make([]services.Endpoint, 0)

	podIP := pod.Status.PodIP
	if pod.Status.Phase != runningPhase {
		return nil
	}

	if len(podIP) == 0 {
		o.logger.WithFields(log.Fields{
			"podName": pod.Name,
		}).Warn("Pod does not have an IP Address")
		return nil
	}

	annotationConfs := annotationConfigsForPod(pod, portAnnotationSet)

	orchestration := services.NewOrchestration("kubernetes", services.KUBERNETES, services.PRIVATE)

	portsSeen := map[int32]bool{}

	podDims := map[string]string{
		"kubernetes_pod_name":  pod.Name,
		"kubernetes_pod_uid":   string(pod.UID),
		"kubernetes_namespace": pod.Namespace,
	}

	makeBaseEndpoint := func(idSuffix string, name string) *services.EndpointCore {
		id := fmt.Sprintf("%s-%s-%s", pod.Name, pod.UID[:7], idSuffix)

		endpoint := services.NewEndpointCore(id, name, observerType, podDims)

		endpoint.Host = podIP

		endpoint.AddExtraField("kubernetes_annotations", pod.Annotations)
		endpoint.AddExtraField("pod_metadata", &pod.ObjectMeta)
		endpoint.AddExtraField("pod_spec", &pod.Spec)

		return endpoint
	}

	for _, container := range pod.Spec.Containers {
		var containerState string
		var containerID string
		var containerName string

		for _, status := range pod.Status.ContainerStatuses {
			if container.Name != status.Name {
				continue
			}

			if status.State.Running == nil {
				break
			}
			containerState = "running"
			containerID = k8sutil.StripContainerID(status.ContainerID)
			containerName = status.Name
		}

		if containerState != "running" {
			continue
		}

		endpointContainer := &services.Container{
			ID:      containerID,
			Names:   []string{containerName},
			Image:   container.Image,
			Command: "",
			State:   containerState,
			Labels:  pod.Labels,
		}

		for _, port := range container.Ports {
			portsSeen[port.ContainerPort] = true

			endpoint := makeBaseEndpoint(fmt.Sprintf("%d", port.ContainerPort), port.Name)

			endpoint.AddDimension("container_spec_name", container.Name)

			portAnnotations := annotationConfs.FilterByPortOrPortName(port.ContainerPort, port.Name)
			monitorType, extraConf, err := configFromAnnotations(container.Name, portAnnotations, pod, client)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("K8s port has invalid config annotations")
			} else {
				endpoint.Configuration = extraConf
				endpoint.MonitorType = monitorType
			}

			endpoint.PortType = services.PortType(port.Protocol)
			endpoint.Port = uint16(port.ContainerPort)
			endpoint.Target = services.TargetTypeHostPort

			endpoints = append(endpoints, &services.ContainerEndpoint{
				EndpointCore:  *endpoint,
				AltPort:       0,
				Container:     *endpointContainer,
				Orchestration: *orchestration,
			})
		}
	}

	// Cover all non-declared ports that were specified in annotations
	for portNum, acs := range annotationConfs.GroupByPortNumber() {
		if portsSeen[portNum] {
			// This would have been handled in the above loop.
			continue
		}

		endpoint := makeBaseEndpoint(fmt.Sprintf("%d", portNum), "")

		monitorType, extraConf, err := configFromAnnotations("", acs, pod, client)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("K8s port has invalid config annotations")
		} else {
			endpoint.Configuration = extraConf
			endpoint.MonitorType = monitorType
		}
		endpoint.PortType = services.UNKNOWN
		endpoint.Port = uint16(portNum)
		endpoint.Target = services.TargetTypeHostPort

		endpoints = append(endpoints, &services.ContainerEndpoint{
			EndpointCore:  *endpoint,
			AltPort:       0,
			Orchestration: *orchestration,
		})
	}

	// Create a "port-less" endpoint for the entire pod
	endpoint := makeBaseEndpoint("pod", pod.Name)
	endpoints = append(endpoints, NewPodEndpoint(endpoint, orchestration))

	return endpoints
}

// Shutdown the service differ routine
func (o *Observer) Shutdown() {
	o.stopIfRunning()
}
