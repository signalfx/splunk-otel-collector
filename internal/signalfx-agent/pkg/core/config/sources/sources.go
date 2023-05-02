// Package sources contains all of the config source logic.  This includes
// logic to get config content from various sources such as the filesystem or a
// KV store.
// It also contains the logic for filling in dynamic values in config.
package sources

import (
	"fmt"
	"reflect"
	"time"

	"github.com/mitchellh/hashstructure"
	"github.com/signalfx/defaults"
	"github.com/signalfx/signalfx-agent/pkg/core/config/sources/consul"
	"github.com/signalfx/signalfx-agent/pkg/core/config/sources/env"
	"github.com/signalfx/signalfx-agent/pkg/core/config/sources/etcd2"
	"github.com/signalfx/signalfx-agent/pkg/core/config/sources/file"
	"github.com/signalfx/signalfx-agent/pkg/core/config/sources/vault"
	"github.com/signalfx/signalfx-agent/pkg/core/config/sources/zookeeper"
	"github.com/signalfx/signalfx-agent/pkg/core/config/types"
	"github.com/signalfx/signalfx-agent/pkg/core/config/validation"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	log "github.com/sirupsen/logrus"

	yaml "gopkg.in/yaml.v2"
)

// SourceConfig represents configuration for various config sources that we
// support.
type SourceConfig struct {
	// Whether to watch config sources for changes.  If this is `true` and
	// the main agent.yaml changes, the agent will dynamically reconfigure
	// itself with minimal disruption.
	// This is generally better than restarting the agent on
	// config changes since that can result in larger gaps in metric data.  The
	// main disadvantage of watching is slightly greater network and compute
	// resource usage. This option is not itself watched for changes. If you
	// change the value of this option, you must restart the agent.
	Watch *bool `yaml:"watch" default:"true"`
	// Whether to watch remote config sources for changes.  If this is `true`
	// and the remote configs changes, the agent will dynamically reconfigure
	// itself with minimal disruption.
	// This is generally better than restarting the agent on
	// config changes since that can result in larger gaps in metric data.  The
	// main disadvantage of watching is slightly greater network and compute
	// resource usage. This option is not itself watched for changes. If you
	// change the value of this option, you must restart the agent.
	RemoteWatch *bool `yaml:"remoteWatch" default:"true"`
	// Configuration for other file sources
	File file.Config `yaml:"file" default:"{}"`
	// Configuration for a Zookeeper remote config source
	Zookeeper *zookeeper.Config `yaml:"zookeeper"`
	// Configuration for an Etcd 2 remote config source
	Etcd2 *etcd2.Config `yaml:"etcd2"`
	// Configuration for a Consul remote config source
	Consul *consul.Config `yaml:"consul"`
	// Configuration for a Hashicorp Vault remote config source
	Vault *vault.Config `yaml:"vault"`
}

// Hash calculates a unique hash value for this config struct
func (sc *SourceConfig) Hash() uint64 {
	hash, err := hashstructure.Hash(sc, nil)
	if err != nil {
		log.WithError(err).Error("Could not get hash of SourceConfig struct")
		return 0
	}
	return hash
}

// SourceInstances returns a map of instantiated sources based on the config
func (sc *SourceConfig) SourceInstances() (map[string]types.ConfigSource, error) {
	sources := make(map[string]types.ConfigSource)

	file := file.New(time.Duration(sc.File.PollRateSeconds) * time.Second)
	sources[file.Name()] = file

	env := env.New()
	sources[env.Name()] = env

	for _, csc := range []types.ConfigSourceConfig{
		sc.Zookeeper,
		sc.Etcd2,
		sc.Consul,
		sc.Vault,
	} {
		if !reflect.ValueOf(csc).IsNil() {
			err := defaults.Set(csc)
			if err != nil {
				panic("Could not set default on source config: " + err.Error())
			}

			if err := validation.ValidateStruct(csc); err != nil {
				return nil, fmt.Errorf("error validating remote config sources: %w", err)
			}

			if err := csc.Validate(); err != nil {
				return nil, fmt.Errorf("error validating remote config sources: %w", err)
			}

			s, err := csc.New()
			if err != nil {
				return nil, fmt.Errorf("error initializing remote config source: %w", err)
			}
			sources[s.Name()] = s
		}
	}
	return sources, nil
}

func parseSourceConfig(config []byte) (SourceConfig, error) {
	var out struct {
		Sources SourceConfig `yaml:"configSources"`
	}

	err := yaml.Unmarshal(config, &out)
	if err != nil {
		return out.Sources, utils.YAMLErrorWithContext(config, err)
	}

	err = defaults.Set(&out.Sources)
	if err != nil {
		panic("Could not set default on source config: " + err.Error())
	}

	return out.Sources, nil
}

// ReadConfig reads in the main agent config file and optionally watches for
// changes on it.  It will be returned immediately, along with a channel that
// will be sent any updated config content if watching is enabled.
func ReadConfig(configPath string, stop <-chan struct{}) ([]byte, <-chan []byte, error) {
	// Fetch the config file with a dummy file source since we don't know what
	// poll rate to configure on it yet.
	contentMap, version, err := file.New(1 * time.Second).Get(configPath)
	if err != nil {
		return nil, nil, err
	}
	if len(contentMap) > 1 {
		return nil, nil, fmt.Errorf("path %s resulted in multiple files", configPath)
	}
	if len(contentMap) == 0 {
		return nil, nil, fmt.Errorf("config file %s could not be found", configPath)
	}

	configContent := contentMap[configPath]
	sourceConfig, err := parseSourceConfig(configContent)
	if err != nil {
		return nil, nil, err
	}

	// Now that we know the poll rate for files, we can make a new file source
	// that will be used for the duration of the agent process.
	fileSource := file.New(time.Duration(sourceConfig.File.PollRateSeconds) * time.Second)

	if *sourceConfig.Watch {
		log.Info("Watching for config file changes")
		changes := make(chan []byte)
		go func() {
			for {
				err := fileSource.WaitForChange(configPath, version, stop)

				if utils.IsSignalChanClosed(stop) {
					return
				}

				if err != nil {
					log.WithError(err).Error("Could not wait for changes to config file")
					time.Sleep(5 * time.Second)
					continue
				}

				log.Info("Config file changed")

				contentMap, version, err = fileSource.Get(configPath)
				if err != nil {
					log.WithError(err).Error("Could not get config file after it was changed")
					time.Sleep(5 * time.Second)
					continue
				}

				changes <- contentMap[configPath]
			}
		}()
		return configContent, changes, nil
	}

	return configContent, nil, nil
}

// DynamicValueProvider handles setting up and providing dynamic values from
// remote config sources.
type DynamicValueProvider struct {
	lastRemoteConfigSourceHash uint64
	sources                    map[string]types.ConfigSource
}

// ReadDynamicValues takes the config file content and processes it for any
// dynamic values of the form `{"#from": ...`.  It returns a YAML document that
// contains the rendered values.  It will optionally watch the sources of any
// dynamic values configured and send updated YAML docs on the returned
// channel.
func (dvp *DynamicValueProvider) ReadDynamicValues(configContent []byte, stop <-chan struct{}) ([]byte, <-chan []byte, error) {
	sourceConfig, err := parseSourceConfig(configContent)
	if err != nil {
		return nil, nil, err
	}

	hash := sourceConfig.Hash()
	if hash != dvp.lastRemoteConfigSourceHash {
		for name, source := range dvp.sources {
			if stoppable, ok := source.(types.Stoppable); ok {
				log.Infof("Stopping stale %s remote config source", name)
				if err := stoppable.Stop(); err != nil {
					log.WithError(err).Errorf("Could not stop stale %s remote config source", name)
				}
			}
		}
		dvp.sources, err = sourceConfig.SourceInstances()
		if err != nil {
			return nil, nil, err
		}
		dvp.lastRemoteConfigSourceHash = hash
	}

	// This is what the cacher will notify on with the names of dynamic value
	// paths that change
	pathChanges := make(chan string)

	cachers := make(map[string]*configSourceCacher)
	for name, source := range dvp.sources {
		cacher := newConfigSourceCacher(source, pathChanges, stop, *sourceConfig.RemoteWatch)
		cachers[name] = cacher
	}

	resolver := newResolver(cachers)

	renderedContent, err := renderDynamicValues(configContent, resolver.Resolve)
	if err != nil {
		return nil, nil, err
	}

	var changes chan []byte
	if *sourceConfig.RemoteWatch {
		changes = make(chan []byte)

		go func() {
			for {
				select {
				case path := <-pathChanges:
					log.Debugf("Dynamic value path %s changed", path)

					renderedContent, err = renderDynamicValues(configContent, resolver.Resolve)
					if err != nil {
						log.WithError(err).Error("Could not render dynamic values in config after change")
						time.Sleep(5 * time.Second)
						continue
					}

					changes <- renderedContent
				case <-stop:
					return
				}
			}
		}()
	}

	return renderedContent, changes, nil
}
