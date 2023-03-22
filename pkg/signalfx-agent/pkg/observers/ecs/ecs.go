// Package ecs is a file-based observer that is primarily meant for
// development and test purposes.  It will watch a json file, which should
// consist of an array of serialized service instances.
package ecs

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/common/ecs"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/core/services"
	"github.com/signalfx/signalfx-agent/pkg/observers"
	"github.com/signalfx/signalfx-agent/pkg/observers/docker"
)

const (
	observerType = "ecs"
)

// OBSERVER(ecs): Queries the [ECS Task Metadata Endpoint version 2](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-metadata-endpoint-v2.html) for running containers.
//
// The Smart Agent needs to run in the same task with the containers to be
// monitored to get access to the data through ECS task metadata endpoint.
//
// ## Configuration from Labels
// You must put at least one special label to identify an exposed port on your Docker
// containers in the ECS task definition since ECS metadata does not contain any port
// information. You can either specify all of the configuration in container labels,
// or you can use the more traditional agent configuration with discovery rules and
// specify configuration overrides with labels.
//
// The config labels are of the form `agent.signalfx.com.config.<port
// number>.<config_key>: <config value>`.  The `<config value>` must be a
// string in a container label, but it will be deserialized as a YAML value to
// the most appropriate type when consumed by the agent.  For example, if you
// have a Redis container and want to monitor it at a higher frequency than
// other Redis containers, you could have an agent config that looks like the
// following:
//
// ```
// observers:
//  - type: ecs
// monitors:
//  - type: collectd/redis
//    discoveryRule: container_image =~ "redis" && port == 6379
//    auth: mypassword
//    intervalSeconds: 10
// ```
//
// And then launch the Redis container with the label:
//
// `agent.signalfx.com.config.6379.intervalSeconds`: `1`
//
// This would cause the config value for `intervalSeconds` to be overwritten to
// the more frequent 1 second interval.
//
// You can also specify the monitor configuration entirely with Docker labels
// and completely omit monitor config from the agent config.  With the agent
// config:
//
// ```
// observers:
//  - type: ecs
// ```
//
// You can then launch a Redis container with the following labels:
//
//  - `agent.signalfx.com.monitorType.6379`: `collectd/redis`
//  - `agent.signalfx.com.config.6379.auth`: `mypassword`
//
// Which would configure a Redis monitor with the given authentication
// configuration.  No Redis configuration is required in the agent config file.
//
// The distinction is that the `monitorType` label was added to the Docker
// container.  If a `monitorType` label is present, **no discovery rules will
// be considered for this endpoint**, and thus, no agent configuration can be
// used anyway.
//
// ### Multiple Monitors per Port
// If you want to configure multiple monitors per port, you can specify the
// port name in the form `<port number>-<port name>` instead of just the port
// number.  For example, if you had two different Prometheus exporters running
// on the same port, but on different paths in a given container, you could
// provide labels like the following:
//
// ```
//  - `agent.signalfx.com.monitorType.8080-app`: `prometheus-exporter`
//  - `agent.signalfx.com.config.8080-app.metricPath`: `/appMetrics`
//  - `agent.signalfx.com.monitorType.8080-goruntime`: `prometheus-exporter`
//  - `agent.signalfx.com.config.8080-goruntime.metricPath`: `/goMetrics`
// ```
//
// The name that is given to the port will populate the `name` field of the
// discovered endpoint and can be used in discovery rules as such.  For
// example, with the following agent config:
//
// ```
// observers:
//  - type: ecs
// monitors:
//  - type: prometheus-exporter
//    discoveryRule: name == "app" && port == 8080
//    intervalSeconds: 1
// ```
//
// And given docker labels as follows (remember that discovery rules are
// irrelevant to endpoints that specify `monitorType` labels):
//
//  - `agent.signalfx.com.config.8080-app.metricPath`: `/appMetrics`
//  - `agent.signalfx.com.config.8080-goruntime.metricPath`: `/goMetrics`
//
// Would result in the `app` endpoint getting an interval of 1 second and the
// `goruntime` endpoint getting the default interval of the agent.

// ENDPOINT_TYPE(ContainerEndpoint): true

var logger = log.WithFields(log.Fields{"observerType": observerType})

// Config for the ecs observer
type Config struct {
	config.ObserverConfig
	// The URL of the ECS task metadata. Default is http://169.254.170.2/v2/metadata, which is hardcoded by AWS for version 2.
	MetadataEndpoint string `yaml:"metadataEndpoint" default:"http://169.254.170.2/v2/metadata"`
	// A mapping of container label names to dimension names that will get
	// applied to the metrics of all discovered services. The corresponding
	// label values will become the dimension values for the mapped name.  E.g.
	// `io.kubernetes.container.name: container_spec_name` would result in a
	// dimension called `container_spec_name` that has the value of the
	// `io.kubernetes.container.name` container label.
	LabelsToDimensions map[string]string `yaml:"labelsToDimensions"`
}

// ECS observer plugin
type ECS struct {
	serviceCallbacks *observers.ServiceCallbacks
	serviceDiffer    *observers.ServiceDiffer
	config           *Config
}

func init() {
	observers.Register(observerType, func(cbs *observers.ServiceCallbacks) interface{} {
		return &ECS{
			serviceCallbacks: cbs,
		}
	}, &Config{})
}

// Configure the ecs observer
func (o *ECS) Configure(config *Config) error {
	o.config = config

	if o.serviceDiffer != nil {
		o.serviceDiffer.Stop()
	}

	o.serviceDiffer = &observers.ServiceDiffer{
		DiscoveryFn:     o.discover,
		IntervalSeconds: 5,
		Callbacks:       o.serviceCallbacks,
	}
	o.serviceDiffer.Start()

	return nil
}

// Discover services from ECS task metadata endpoint
func (o *ECS) discover() []services.Endpoint {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	response, err := client.Get(o.config.MetadataEndpoint)
	if err != nil {
		logger.WithError(err).Error("Could not connect to metadata endpoint")
		return nil
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		logger.WithError(err).Errorf("Could not receive ECS metadata : %s", http.StatusText(response.StatusCode))
		return nil
	}

	var metadata ecs.TaskMetadata

	if err := json.NewDecoder(response.Body).Decode(&metadata); err != nil {
		logger.WithFields(log.Fields{
			"error": err,
		}).Error("Could not parse metadata json")
		return nil
	}

	var out []services.Endpoint

	for i := range metadata.Containers {
		endpoints := o.endpointsForContainer(&metadata.Containers[i], metadata.GetDimensions())
		out = append(out, endpoints...)
	}

	return out
}

func (o *ECS) endpointsForContainer(cont *ecs.Container, taskDims map[string]string) []services.Endpoint {
	instances := make([]services.Endpoint, 0)

	if cont.KnownStatus == "RUNNING" {
		labelConfigs := docker.GetConfigLabels(cont.Labels)
		knownPorts := map[docker.ContPort]bool{}

		for port := range labelConfigs {
			knownPorts[port] = true
		}

		for portObj := range knownPorts {
			endpoint := o.endpointForPort(portObj, cont, taskDims)

			if labelConf := labelConfigs[portObj]; labelConf != nil {
				endpoint.MonitorType = labelConf.MonitorType
				endpoint.Configuration = labelConf.Configuration
			}

			instances = append(instances, endpoint)
		}

		// Add an "port-less" endpoint that identifies the container in
		// general.
		containerEndpoint := o.makeBaseEndpointForContainer(
			cont,
			"portless",
			strings.TrimLeft(cont.Name, "/"),
			taskDims)
		containerEndpoint.Target = services.TargetTypeContainer
		instances = append(instances, containerEndpoint)
	}

	return instances
}

func (o *ECS) makeBaseEndpointForContainer(cont *ecs.Container, idSuffix, name string, taskDims map[string]string) *services.ContainerEndpoint {
	hostIP := cont.Networks[0].IPAddresses[0]

	serviceContainer := &services.Container{
		ID:     cont.DockerID,
		Names:  []string{cont.Name},
		Image:  cont.Image,
		State:  cont.KnownStatus,
		Labels: cont.Labels,
	}

	orchDims := map[string]string{}
	for dimName, v := range taskDims {
		orchDims[dimName] = v
	}
	for k, dimName := range o.config.LabelsToDimensions {
		if v := cont.Labels[k]; v != "" {
			orchDims[dimName] = v
		}
	}

	id := serviceContainer.PrimaryName() + "-" + serviceContainer.ID[:12] + "-" + idSuffix
	endpoint := &services.ContainerEndpoint{
		EndpointCore:  *services.NewEndpointCore(id, name, observerType, orchDims),
		Container:     *serviceContainer,
		Orchestration: *services.NewOrchestration("ecs", services.ECS, services.PRIVATE),
	}
	endpoint.Host = hostIP

	return endpoint
}

func (o *ECS) endpointForPort(portObj docker.ContPort, cont *ecs.Container, taskDims map[string]string) *services.ContainerEndpoint {
	port := portObj.Int()
	protocol := portObj.Proto()

	idSuffix := strconv.Itoa(port)
	if portObj.Name != "" {
		idSuffix += "-" + portObj.Name
	}

	endpoint := o.makeBaseEndpointForContainer(cont, idSuffix, portObj.Name, taskDims)

	endpoint.Port = uint16(port)
	endpoint.AltPort = uint16(port)
	endpoint.PortType = services.PortType(strings.ToUpper(protocol))
	endpoint.Target = services.TargetTypeHostPort

	return endpoint
}

// Shutdown the service differ routine
func (o *ECS) Shutdown() {
	if o.serviceDiffer != nil {
		o.serviceDiffer.Stop()
	}
}
