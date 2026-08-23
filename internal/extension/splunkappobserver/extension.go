// Copyright Splunk, Inc.
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

package splunkappobserver

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/endpointswatcher"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
	"go.uber.org/zap"
)

const TypeStr = "splunk_app_observer"

var (
	_ extension.Extension      = (*splunkAppObserver)(nil)
	_ observer.Observable      = (*splunkAppObserver)(nil)
	_ observer.EndpointDetails = (*splunkAppDetails)(nil)

	errMissingSplunkHome = errors.New("SPLUNK_HOME is not set")
)

type splunkAppObserver struct {
	*endpointswatcher.EndpointsWatcher
}

type endpointsLister struct {
	logger       *zap.Logger
	observerName string
	appsPath     string
	readDir      func(string) ([]os.DirEntry, error)
	stat         func(string) (os.FileInfo, error)
}

type splunkAppDetails struct {
	Name string
	Path string
}

func (a *splunkAppDetails) Env() observer.EndpointEnv {
	return observer.EndpointEnv{
		"name": a.Name,
		"path": a.Path,
	}
}

func (*splunkAppDetails) Type() observer.EndpointType {
	// TODO: Resolve this upstream by better extensibility support on receiver_create
	return observer.EndpointType("splunk_app")
}

func newObserver(settings extension.Settings, config *Config) extension.Extension {
	lister := &endpointsLister{
		logger:       settings.Logger,
		observerName: settings.ID.String(),
		appsPath:     config.AppsPath,
		readDir:      os.ReadDir,
		stat:         os.Stat,
	}

	return &splunkAppObserver{
		EndpointsWatcher: endpointswatcher.New(lister, config.RefreshInterval, settings.Logger),
	}
}

func (*splunkAppObserver) Start(context.Context, component.Host) error {
	return nil
}

func (o *splunkAppObserver) Shutdown(context.Context) error {
	o.StopListAndWatch()
	return nil
}

func (l *endpointsLister) ListEndpoints() []observer.Endpoint {
	appsPath, err := resolveAppsPath(l.appsPath)
	if err != nil {
		l.logger.Warn("Could not resolve Splunk apps path", zap.String("apps_path", l.appsPath), zap.Error(err))
		return nil
	}

	entries, err := l.readDir(appsPath)
	if err != nil {
		l.logger.Warn("Could not list Splunk app directories", zap.String("apps_path", appsPath), zap.Error(err))
		return nil
	}

	endpoints := make([]observer.Endpoint, 0, len(entries))
	for _, entry := range entries {
		if !l.isDir(appsPath, entry) {
			continue
		}

		appName := entry.Name()
		pathToApp := filepath.Join(appsPath, appName)
		appPath, err := filepath.Abs(pathToApp)
		if err != nil {
			l.logger.Error("Could not obtain the absolute path to the splunk app", zap.String("path", pathToApp), zap.Error(err))
			continue
		}
		endpoints = append(endpoints, observer.Endpoint{
			ID:      observer.EndpointID(fmt.Sprintf("%s/%s", l.observerName, appName)),
			Target:  appPath,
			Details: &splunkAppDetails{Name: appName, Path: appPath},
		})
	}
	return endpoints
}

func (l *endpointsLister) isDir(appsPath string, entry os.DirEntry) bool {
	if entry.IsDir() {
		return true
	}
	if entry.Type()&fs.ModeSymlink == 0 {
		return false
	}

	path := filepath.Join(appsPath, entry.Name())
	info, err := l.stat(path)
	if err != nil {
		l.logger.Warn("Could not inspect Splunk app path", zap.String("path", path), zap.Error(err))
		return false
	}
	return info.IsDir()
}

func resolveAppsPath(path string) (string, error) {
	if path == "" {
		return "", errors.New("apps_path is empty")
	}
	if referencesSplunkHome(path) && os.Getenv("SPLUNK_HOME") == "" {
		return "", errMissingSplunkHome
	}
	return filepath.Clean(os.ExpandEnv(path)), nil
}

func referencesSplunkHome(path string) bool {
	return path == "$SPLUNK_HOME" ||
		path == "${SPLUNK_HOME}" ||
		strings.HasPrefix(path, "$SPLUNK_HOME/") ||
		strings.HasPrefix(path, "$SPLUNK_HOME\\") ||
		strings.HasPrefix(path, "${SPLUNK_HOME}/") ||
		strings.HasPrefix(path, "${SPLUNK_HOME}\\")
}
