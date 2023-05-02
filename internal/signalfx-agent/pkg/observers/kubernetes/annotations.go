package kubernetes

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	k8s "k8s.io/client-go/kubernetes"

	"github.com/signalfx/signalfx-agent/pkg/utils"
	"github.com/signalfx/signalfx-agent/pkg/utils/k8sutil"
)

var annotationConfigRegexp = regexp.MustCompile(
	`^agent.signalfx.com/` +
		`(?P<type>monitorType|config|configFromEnv|configFromSecret)` +
		`.(?P<port>[\w-]+)` +
		`(?:.(?P<config_key>\w+))?$`)

// AnnotationConfig is a generic struct that can describe any of the annotation
// config values we support.
type AnnotationConfig struct {
	AnnotationKey string
	// The type of annotation
	Type string
	// Either the port number or name must be specified
	Port     int32
	PortName string
	// The config key that this will result in when configuring a monitor
	ConfigKey string
	Value     string
}

// AnnotationConfigs is a slice of AnnotationConfig with some helper methods
// for filtering.
type AnnotationConfigs []*AnnotationConfig

// FilterByPortOrPortName returns all AnnotationConfig instances that match
// either the port number or port name.
func (ac AnnotationConfigs) FilterByPortOrPortName(port int32, portName string) (out AnnotationConfigs) {
	for i := range ac {
		if ac[i].Port == port || (portName != "" && ac[i].PortName == portName) {
			out = append(out, ac[i])
		}
	}
	return
}

func (ac AnnotationConfigs) GroupByPortNumber() map[int32]AnnotationConfigs {
	out := map[int32]AnnotationConfigs{}
	for i := range ac {
		port := ac[i].Port
		if port == 0 {
			// Only deal with annotations that specify a port by number
			continue
		}

		out[port] = append(out[port], ac[i])
	}
	return out
}

func parseAgentAnnotation(key, value string, pod *v1.Pod) (*AnnotationConfig, error) {
	groups := annotationConfigRegexp.FindStringSubmatch(key)
	if len(groups) == 0 || groups[0] == "" {
		return nil, fmt.Errorf("kubernetes config annotation has invalid agent namespaced key: %s", key)
	}

	conf := &AnnotationConfig{
		AnnotationKey: key,
		Type:          groups[1],
		ConfigKey:     groups[3],
		Value:         value,
	}

	portStr := groups[2]
	if portInt, err := strconv.Atoi(portStr); err != nil {
		conf.PortName = portStr
	} else {
		conf.Port = int32(portInt)
	}

	if conf.Type != "monitorType" && len(conf.ConfigKey) == 0 {
		return nil, fmt.Errorf("kubernetes config annotation %s is missing a config key", key)
	}
	if conf.Port != 0 && k8sutil.PortByNumber(pod, conf.Port) == nil {
		return nil, fmt.Errorf("kubernetes config annotation %s references invalid port number %d", key, conf.Port)
	}
	if conf.PortName != "" && k8sutil.PortByName(pod, conf.PortName) == nil {
		return nil, fmt.Errorf("kubernetes config annotation %s references invalid port name %s", key, conf.PortName)
	}

	return conf, nil
}

func annotationConfigsForPod(pod *v1.Pod, additionalPortAnnotations map[string]bool) AnnotationConfigs {
	var confs []*AnnotationConfig

	for key, value := range pod.Annotations {
		var annotationConf *AnnotationConfig
		if strings.HasPrefix(key, "agent.signalfx.com") {
			var err error
			annotationConf, err = parseAgentAnnotation(key, value, pod)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("Invalid K8s agent annotation")
				continue
			}
		} else if additionalPortAnnotations[key] {
			port, err := strconv.Atoi(value)
			if err != nil {
				log.WithField("annotation", key).WithField("value", value).Error("K8s annotation should be a port number")
				continue
			}
			annotationConf = &AnnotationConfig{
				AnnotationKey: key,
				Port:          int32(port),
			}
		}

		if annotationConf == nil {
			continue
		}

		confs = append(confs, annotationConf)
	}

	return AnnotationConfigs(confs)
}

func configFromAnnotations(
	container string, annotationConfs AnnotationConfigs, pod *v1.Pod, client *k8s.Clientset) (string, map[string]interface{}, error) {

	extraConfig := make(map[string]interface{})
	var monitorType string

	for _, ac := range annotationConfs {
		switch ac.Type {
		case "monitorType":
			monitorType = ac.Value

		case "config":
			extraConfig[ac.ConfigKey] = utils.DecodeValueGenerically(strings.TrimSpace(ac.Value))

		case "configFromEnv":
			val, err := k8sutil.EnvValueForContainer(pod, ac.Value, container)
			if err != nil {
				return "", nil, err
			}
			extraConfig[ac.ConfigKey] = utils.DecodeValueGenerically(strings.TrimSpace(val))

		case "configFromSecret":
			parts := strings.SplitN(ac.Value, "/", 2)
			if len(parts) != 2 {
				return "", nil, fmt.Errorf("%s value '%s' should be of the form <secretName>/<dataKey>", ac.AnnotationKey, ac.Value)
			}

			secret, err := k8sutil.FetchSecretValue(client, parts[0], parts[1], pod.Namespace)
			if err != nil {
				return "", nil, fmt.Errorf("could not fetch k8s secret: %w", err)
			}
			// Always treat secret values as strings
			extraConfig[ac.ConfigKey] = secret
		}
	}

	return monitorType, extraConfig, nil
}
