// Copyright Splunk, Inc.
// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package configconverter

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cast"
	"go.opentelemetry.io/collector/confmap"
	"gopkg.in/yaml.v2"
)

const (
	configServerEnabledEnvVar   = "SPLUNK_DEBUG_CONFIG_SERVER"
	configServerPortEnvVar      = "SPLUNK_DEBUG_CONFIG_SERVER_PORT"
	defaultConfigServerEndpoint = "localhost:55554"
	effectivePath               = "/debug/configz/effective"
	initialPath                 = "/debug/configz/initial"
)

type ConfigType int

const (
	initialConfig   ConfigType = 1
	effectiveConfig ConfigType = 2
)

var _ confmap.Converter = (*ConfigServer)(nil)

type ConfigServer struct {
	// Use get/set methods instead of direct usage
	initial        map[string]any
	effective      map[string]any
	server         *http.Server
	doneCh         chan struct{}
	initialMutex   sync.RWMutex
	effectiveMutex sync.RWMutex
	wg             sync.WaitGroup
	once           sync.Once
}

func NewConfigServer() *ConfigServer {
	cs := &ConfigServer{
		initial:        map[string]any{},
		effective:      map[string]any{},
		initialMutex:   sync.RWMutex{},
		effectiveMutex: sync.RWMutex{},
		wg:             sync.WaitGroup{},
		once:           sync.Once{},
		doneCh:         make(chan struct{}),
	}

	mux := http.NewServeMux()
	initialHandleFunc := cs.muxHandleFunc(initialConfig)
	mux.HandleFunc(initialPath, initialHandleFunc)

	effectiveHandleFunc := cs.muxHandleFunc(effectiveConfig)
	mux.HandleFunc(effectivePath, effectiveHandleFunc)

	cs.server = &http.Server{
		ReadHeaderTimeout: 20 * time.Second,
		Handler:           mux,
	}
	return cs
}

// Convert is intended to be called as the final service confmap.Converter,
// which registers the service config before being finally resolved and unmarshalled.
func (cs *ConfigServer) Convert(_ context.Context, conf *confmap.Conf) error {
	cs.start()
	cs.setEffective(conf.ToStringMap())
	return nil
}

func (cs *ConfigServer) Register() {
	cs.wg.Add(1)
}

func (cs *ConfigServer) SetForScheme(scheme string, config map[string]any) {
	cs.initialMutex.Lock()
	defer cs.initialMutex.Unlock()
	cs.initial[scheme] = config
}

func (cs *ConfigServer) getInitial() map[string]any {
	cs.initialMutex.RLock()
	defer cs.initialMutex.RUnlock()
	return cs.initial
}

func (cs *ConfigServer) setEffective(config map[string]any) {
	cs.effectiveMutex.Lock()
	defer cs.effectiveMutex.Unlock()
	cs.effective = config
}

func (cs *ConfigServer) getEffective() map[string]any {
	cs.effectiveMutex.RLock()
	defer cs.effectiveMutex.RUnlock()
	return cs.effective
}

// start will create and start the singleton http server. It presumes cs.Register() has been
// called at least once and will tear down the moment the final cs.Unregister() call is made.
func (cs *ConfigServer) start() {
	if enabled := os.Getenv(configServerEnabledEnvVar); enabled != "true" {
		// The config server needs to be explicitly enabled for the time being.
		return
	}

	cs.once.Do(
		func() {
			endpoint := defaultConfigServerEndpoint
			if portOverride, ok := os.LookupEnv(configServerPortEnvVar); ok {
				if portOverride == "" {
					// If explicitly set to empty do not start the server.
					return
				}

				endpoint = "localhost:" + portOverride
			}

			listener, err := net.Listen("tcp", endpoint)
			if err != nil {
				if errors.Is(err, syscall.EADDRINUSE) {
					err = fmt.Errorf("%w: please set %q environment variable to nonconflicting port", err, configServerPortEnvVar)
				}
				log.Print(fmt.Errorf("error starting config server: %w", err).Error())
				return
			}

			go func() {
				defer close(cs.doneCh)

				httpErr := cs.server.Serve(listener)
				if httpErr != http.ErrServerClosed {
					log.Print(fmt.Errorf("config server error: %w", httpErr).Error())
				}
			}()

			go func() {
				cs.wg.Wait()
				if cs.server != nil {
					_ = cs.server.Close()
					// If launched wait for Serve goroutine exit.
					<-cs.doneCh
				}

			}()

		})
}

func (cs *ConfigServer) Unregister() {
	cs.wg.Done()
}

func (cs *ConfigServer) muxHandleFunc(configType ConfigType) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "GET" {
			writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var configYAML []byte
		if configType == initialConfig {
			configYAML, _ = yaml.Marshal(cs.getInitial())
		} else {
			configYAML, _ = yaml.Marshal(simpleRedact(cs.getEffective()))
		}
		_, _ = writer.Write(configYAML)
	}
}

func simpleRedact(config map[string]any) map[string]any {
	redactedConfig := make(map[string]any)
	for k, v := range config {
		switch value := v.(type) {
		case string:
			if shouldRedactKey(k) {
				v = "<redacted>"
			}
		case map[string]any:
			v = simpleRedact(value)
		case map[any]any:
			v = simpleRedact(cast.ToStringMap(value))
		}

		redactedConfig[k] = v
	}

	return redactedConfig
}

// shouldRedactKey applies a simple check to see if the contents of the given key
// should be redacted or not.
func shouldRedactKey(k string) bool {
	fragments := []string{
		"access",
		"api_key",
		"apikey",
		"auth",
		"credential",
		"creds",
		"login",
		"password",
		"pwd",
		"token",
		"user",
	}

	for _, fragment := range fragments {
		if strings.Contains(k, fragment) {
			return true
		}
	}

	return false
}
