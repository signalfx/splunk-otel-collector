// Copyright Splunk, Inc.
// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package includeconfigsource

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"text/template"

	"github.com/fsnotify/fsnotify"
	"go.opentelemetry.io/collector/config/experimental/configsource"

	"github.com/signalfx/splunk-otel-collector/internal/configprovider"
)

// Private error types to help with testability.
type (
	errFailedToDeleteFile struct{ error }
)

// includeSession implements the configsource.Session interface.
type includeSession struct {
	watcher      *fsnotify.Watcher
	watchedFiles map[string]struct{}
	Config
}

var _ configsource.Session = (*includeSession)(nil)

func (is *includeSession) Retrieve(_ context.Context, selector string, params interface{}) (configsource.Retrieved, error) {
	tmpl, err := template.ParseFiles(selector)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, params); err != nil {
		return nil, err
	}

	if is.DeleteFiles {
		if err = os.Remove(selector); err != nil {
			return nil, &errFailedToDeleteFile{fmt.Errorf("failed to delete file %q as requested: %w", selector, err)}
		}
	}

	if !is.WatchFiles {
		return configprovider.NewRetrieved(buf.Bytes()), nil
	}

	watchForUpdateFn, err := is.watchFile(selector)
	if err != nil {
		return nil, err
	}

	if watchForUpdateFn == nil {
		return configprovider.NewRetrieved(buf.Bytes()), nil
	}
	return configprovider.NewWatchableRetrieved(buf.Bytes(), watchForUpdateFn), nil
}

func (is *includeSession) RetrieveEnd(context.Context) error {
	return nil
}

func (is *includeSession) Close(context.Context) error {
	if is.watcher != nil {
		return is.watcher.Close()
	}

	return nil
}

func newSession(config Config) *includeSession {
	return &includeSession{
		Config:       config,
		watchedFiles: make(map[string]struct{}),
	}
}

func (is *includeSession) watchFile(file string) (func() error, error) {
	var watchForUpdateFn func() error
	if _, watched := is.watchedFiles[file]; watched {
		// This file is already watched another watch function is not needed.
		return watchForUpdateFn, nil
	}

	if is.watcher == nil {
		// First watcher create a real watch for update function.
		var err error
		is.watcher, err = fsnotify.NewWatcher()
		if err != nil {
			return nil, err
		}

		watchForUpdateFn = func() error {
			for {
				select {
				case event, ok := <-is.watcher.Events:
					if !ok {
						return configsource.ErrSessionClosed
					}
					if event.Op&fsnotify.Write == fsnotify.Write {
						return fmt.Errorf("file used in the config modified: %q: %w", event.Name, configsource.ErrValueUpdated)
					}
				case watcherErr, ok := <-is.watcher.Errors:
					if !ok {
						return configsource.ErrSessionClosed
					}
					return watcherErr
				}
			}
		}
	}

	// Now just add the file.
	if err := is.watcher.Add(file); err != nil {
		return nil, err
	}

	is.watchedFiles[file] = struct{}{}

	return watchForUpdateFn, nil
}
