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

package fileconfigsource

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cast"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/experimental/configsource"

	"github.com/signalfx/splunk-otel-collector/internal/configprovider"
)

// Private error types to help with testability.
type (
	errFailedToDeleteFile    struct{ error }
	errInvalidRetrieveParams struct{ error }
	errMissingRequiredFile   struct{ error }
)

type retrieveParams struct {
	// Delete is used to instruct the config source to delete the
	// file after its content is read. The default value is 'false'.
	// Set it to 'true' to force the deletion of the file as soon
	// as the config source finished using it.
	Delete bool `mapstructure:"delete"`
	// DisableWatch is used to control if the referenced file should
	// be watched for updates or not. The default value is 'false'.
	// Set it to 'true' to prevent monitoring for updates on the given
	// file.
	DisableWatch bool `mapstructure:"disable_watch"`
}

// fileSession implements the configsource.Session interface.
type fileSession struct {
	watcher      *fsnotify.Watcher
	watchedFiles map[string]struct{}
}

var _ configsource.Session = (*fileSession)(nil)

func (fs *fileSession) Retrieve(_ context.Context, selector string, params interface{}) (configsource.Retrieved, error) {
	actualParams := retrieveParams{}
	if params != nil {
		paramsParser := config.NewParserFromStringMap(cast.ToStringMap(params))
		if err := paramsParser.UnmarshalExact(&actualParams); err != nil {
			return nil, &errInvalidRetrieveParams{fmt.Errorf("failed to unmarshall retrieve params: %w", err)}
		}
	}

	bytes, err := ioutil.ReadFile(filepath.Clean(selector))
	if err != nil {
		return nil, &errMissingRequiredFile{err}
	}

	if actualParams.Delete {
		if err = os.Remove(selector); err != nil {
			return nil, &errFailedToDeleteFile{fmt.Errorf("failed to delete file %q as requested: %w", selector, err)}
		}
	}

	if actualParams.Delete || actualParams.DisableWatch {
		return configprovider.NewRetrieved(bytes, configprovider.WatcherNotSupported), nil
	}

	watchForUpdateFn, err := fs.watchFile(selector)
	if err != nil {
		return nil, err
	}

	return configprovider.NewRetrieved(bytes, watchForUpdateFn), nil
}

func (fs *fileSession) RetrieveEnd(context.Context) error {
	return nil
}

func (fs *fileSession) Close(context.Context) error {
	if fs.watcher != nil {
		return fs.watcher.Close()
	}

	return nil
}

func newSession() (*fileSession, error) {
	return &fileSession{
		watchedFiles: make(map[string]struct{}),
	}, nil
}

func (fs *fileSession) watchFile(file string) (func() error, error) {
	watchForUpdateFn := configprovider.WatcherNotSupported
	if _, watched := fs.watchedFiles[file]; watched {
		// This file is already watched another watch function is not needed.
		return watchForUpdateFn, nil
	}

	if fs.watcher == nil {
		// First watcher create a real watch for update function.
		var err error
		fs.watcher, err = fsnotify.NewWatcher()
		if err != nil {
			return nil, err
		}

		watchForUpdateFn = func() error {
			for {
				select {
				case event, ok := <-fs.watcher.Events:
					if !ok {
						return configsource.ErrSessionClosed
					}
					if event.Op&fsnotify.Write == fsnotify.Write {
						return fmt.Errorf("file used in the config modified: %q: %w", event.Name, configsource.ErrValueUpdated)
					}
				case watcherErr, ok := <-fs.watcher.Errors:
					if !ok {
						return configsource.ErrSessionClosed
					}
					return watcherErr
				}
			}
		}
	}

	// Now just add the file.
	if err := fs.watcher.Add(file); err != nil {
		return nil, err
	}

	fs.watchedFiles[file] = struct{}{}

	return watchForUpdateFn, nil
}
