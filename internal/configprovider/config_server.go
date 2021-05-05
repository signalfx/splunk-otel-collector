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

package configprovider

import (
	"net"
	"net/http"
	"os"
	"strings"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

const (
	defaultEndpoint = "localhost:5555"
)

type configServer struct {
	logger    *zap.Logger
	initial   map[string]interface{}
	effective map[string]interface{}
	server    *http.Server
	doneCh    chan struct{}
}

func newConfigServer(logger *zap.Logger, initial, effective map[string]interface{}) *configServer {
	return &configServer{
		logger:    logger,
		initial:   initial,
		effective: effective,
	}
}

func (cs *configServer) start() error {
	endpoint := defaultEndpoint
	if portOverride, ok := os.LookupEnv("SPLUNK_CONFIG_SERVER_PORT"); ok {
		if portOverride == "" {
			// If explicitly set to empty do not start the server.
			return nil
		}

		endpoint = "localhost:" + portOverride
	}

	listener, err := net.Listen("tcp", endpoint)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()

	initialHandleFunc, err := cs.muxHandleFunc(cs.initial)
	if err != nil {
		return err
	}
	mux.HandleFunc("/debug/configz/initial", initialHandleFunc)

	effectiveHandleFunc, err := cs.muxHandleFunc(simpleRedact(cs.effective))
	if err != nil {
		return err
	}
	mux.HandleFunc("/debug/configz/effective", effectiveHandleFunc)

	cs.server = &http.Server{
		Handler: mux,
	}
	cs.doneCh = make(chan struct{})
	go func() {
		defer close(cs.doneCh)

		if httpErr := cs.server.Serve(listener); httpErr != http.ErrServerClosed {
			cs.logger.Error("config server error", zap.Error(err))
		}
	}()

	return nil
}

func (cs *configServer) shutdown() error {
	var err error
	if cs.server != nil {
		err = cs.server.Close()
		// If launched wait for Serve goroutine exit.
		<-cs.doneCh
	}

	return err
}

func (cs *configServer) muxHandleFunc(config map[string]interface{}) (func(http.ResponseWriter, *http.Request), error) {
	configYAML, err := yaml.Marshal(config)
	if err != nil {
		return nil, err
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write(configYAML)
	}, nil
}

func simpleRedact(config map[string]interface{}) map[string]interface{} {
	redactedConfig := make(map[string]interface{})
	for k, v := range config {
		switch value := v.(type) {
		case string:
			if shouldRedactKey(k) {
				v = "<redacted>"
			}
		case map[string]interface{}:
			v = simpleRedact(value)
		}

		redactedConfig[k] = v
	}

	return redactedConfig
}

// shouldRedactKey applies a simple check to see if the contents of the given key
// should be redacted or not.
func shouldRedactKey(k string) bool {
	fragments := []string{"access", "auth", "credential", "creds", "login", "password", "pwd", "user"}

	for _, fragment := range fragments {
		if strings.Contains(k, fragment) {
			return true
		}
	}

	return false
}
