package config

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/signalfx/defaults"
	"github.com/signalfx/signalfx-agent/pkg/core/config/sources"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	"github.com/signalfx/signalfx-agent/pkg/utils/structtags"
	log "github.com/sirupsen/logrus"
)

// LoadConfig handles loading the main config file and recursively rendering
// any dynamic values in the config.  If watchInterval is 0, the config will be
// loaded once and sent to the returned channel, after which the channel will
// be closed.  Otherwise, the returned channel will remain open and will be
// sent any config updates.
func LoadConfig(ctx context.Context, configPath string) (<-chan *Config, error) {
	configYAML, configFileChanges, err := sources.ReadConfig(configPath, ctx.Done())
	if err != nil {
		return nil, fmt.Errorf("could not read config file %s: %w", configPath, err)
	}

	dynamicProvider := sources.DynamicValueProvider{}

	dynamicValueCtx, cancelDynamic := context.WithCancel(ctx)
	finalYAML, dynamicChanges, err := dynamicProvider.ReadDynamicValues(configYAML, dynamicValueCtx.Done())
	if err != nil {
		cancelDynamic()
		return nil, err
	}

	config, err := loadYAML(finalYAML)
	if err != nil {
		cancelDynamic()
		return nil, err
	}

	// Give it enough room to hold the initial config load.
	loads := make(chan *Config, 1)

	loads <- config

	if configFileChanges != nil || dynamicChanges != nil {
		go func() {
			for {
				// We can have changes either in the dynamic values or the
				// config file itself.  If the config file changes, we have to
				// recreate the dynamic value watcher since it is configured
				// from the config file.
				select {
				case configYAML = <-configFileChanges:
					cancelDynamic()

					dynamicValueCtx, cancelDynamic = context.WithCancel(ctx)

					finalYAML, dynamicChanges, err = dynamicProvider.ReadDynamicValues(configYAML, dynamicValueCtx.Done())
					if err != nil {
						log.WithError(err).Error("Could not read dynamic values in config after change")
						time.Sleep(5 * time.Second)
						continue
					}

					config, err := loadYAML(finalYAML)
					if err != nil {
						log.WithError(err).Error("Could not parse config after change")
						continue
					}

					loads <- config
				case finalYAML = <-dynamicChanges:
					config, err := loadYAML(finalYAML)
					if err != nil {
						log.WithError(err).Error("Could not parse config after change")
						continue
					}
					loads <- config
				case <-ctx.Done():
					cancelDynamic()
					return
				}
			}
		}()
	} else {
		cancelDynamic()
	}
	return loads, nil
}

func loadYAML(fileContent []byte) (*Config, error) {
	config := &Config{}

	preprocessedContent := preprocessConfig(fileContent)

	err := yaml.UnmarshalStrict(preprocessedContent, config)
	if err != nil {
		return nil, utils.YAMLErrorWithContext(preprocessedContent, err)
	}

	if err := defaults.Set(config); err != nil {
		panic(fmt.Sprintf("Config defaults are wrong types: %s", err))
	}

	if err := structtags.CopyTo(config); err != nil {
		panic(fmt.Sprintf("Error copying configs to fields: %v", err))
	}

	return config.initialize()
}

var envVarRE = regexp.MustCompile(`\${\s*([\w-]+?)\s*}`)

// Hold all of the envvars so that when they are sanitized from the proc we can
// still get to them when we need to rerender config
var envVarCache = make(map[string]string)

var envVarWhitelist = map[string]bool{
	"MY_NODE_NAME": true,
}

// Replaces envvar syntax with the actual envvars
func preprocessConfig(content []byte) []byte {
	return envVarRE.ReplaceAllFunc(content, func(bs []byte) []byte {
		parts := envVarRE.FindSubmatch(bs)
		envvar := string(parts[1])

		val, ok := envVarCache[envvar]

		if !ok {
			val = os.Getenv(envvar)
			envVarCache[envvar] = val

			log.WithFields(log.Fields{
				"envvar": envvar,
			}).Debug("Sanitizing envvar from agent")

			if !envVarWhitelist[envvar] {
				os.Unsetenv(envvar)
			}
		}

		return []byte(val)
	})
}
