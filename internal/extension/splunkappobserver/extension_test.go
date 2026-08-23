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
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/extension/extensiontest"
	"go.uber.org/zap/zaptest"
)

func TestListEndpointsReportsChildDirectories(t *testing.T) {
	appsPath := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(appsPath, "search"), 0o755))
	require.NoError(t, os.Mkdir(filepath.Join(appsPath, "Splunk_TA_example"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(appsPath, "README"), []byte("not an app"), 0o600))

	if runtime.GOOS != "windows" {
		target := t.TempDir()
		require.NoError(t, os.Symlink(target, filepath.Join(appsPath, "symlink_app")))
	}

	lister := &endpointsLister{
		logger:       zaptest.NewLogger(t),
		observerName: TypeStr,
		appsPath:     appsPath,
		readDir:      os.ReadDir,
		stat:         os.Stat,
	}

	endpoints := lister.ListEndpoints()

	expectedNames := []string{"Splunk_TA_example", "search"}
	if runtime.GOOS != "windows" {
		expectedNames = append(expectedNames, "symlink_app")
	}
	require.Len(t, endpoints, len(expectedNames))

	byName := map[string]observer.Endpoint{}
	for _, endpoint := range endpoints {
		env, err := endpoint.Env()
		require.NoError(t, err)
		byName[env["name"].(string)] = endpoint
	}

	for _, name := range expectedNames {
		endpoint, ok := byName[name]
		require.True(t, ok, "missing endpoint for app %q", name)

		appPath := filepath.Join(appsPath, name)
		assert.Equal(t, observer.EndpointID("("+TypeStr+")"+name), endpoint.ID)
		assert.Equal(t, appPath, endpoint.Target)
		assert.Equal(t, observer.ContainerType, endpoint.Details.Type())
		assert.IsType(t, &splunkAppDetails{}, endpoint.Details)

		env, err := endpoint.Env()
		require.NoError(t, err)
		assert.Equal(t, string(observer.ContainerType), env["type"])
		assert.Equal(t, name, env["name"])
		assert.Equal(t, "splunk_app", env["image"])
		assert.Equal(t, appPath, env["command"])
		assert.Equal(t, appPath, env["endpoint"])
		assert.Equal(t, appPath, env["container_id"])
		assert.Equal(t, appPath, env["host"])
		assert.Equal(t, name, env["splunk_app_name"])
		assert.Equal(t, appPath, env["splunk_app_path"])
		assert.Equal(t, map[string]string{
			"splunk_app_name": name,
			"splunk_app_path": appPath,
		}, env["labels"])
	}
}

func TestListAndWatchReportsAddsAndRemoves(t *testing.T) {
	appsPath := t.TempDir()
	cfg := &Config{
		AppsPath:        appsPath,
		RefreshInterval: 10 * time.Millisecond,
	}
	settings := extensiontest.NewNopSettings(component.MustNewType(TypeStr))
	settings.Logger = zaptest.NewLogger(t)

	ext := newObserver(settings, cfg)
	require.NoError(t, ext.Start(t.Context(), componenttest.NewNopHost()))
	t.Cleanup(func() {
		require.NoError(t, ext.Shutdown(t.Context()))
	})

	observable, ok := ext.(observer.Observable)
	require.True(t, ok)

	notify := &recordingNotify{id: "test-notify"}
	observable.ListAndWatch(notify)
	t.Cleanup(func() {
		observable.Unsubscribe(notify)
	})

	appPath := filepath.Join(appsPath, "Splunk_TA_new")
	require.NoError(t, os.Mkdir(appPath, 0o755))
	expectedID := observer.EndpointID("(" + settings.ID.String() + ")Splunk_TA_new")
	require.Eventually(t, func() bool {
		return notify.hasAdded(expectedID)
	}, time.Second, 10*time.Millisecond)

	require.NoError(t, os.Remove(appPath))
	require.Eventually(t, func() bool {
		return notify.hasRemoved(expectedID)
	}, time.Second, 10*time.Millisecond)
}

func TestResolveAppsPath(t *testing.T) {
	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	path, err := resolveAppsPath(defaultAppsPath)

	require.NoError(t, err)
	assert.Equal(t, filepath.Clean("/opt/splunk/etc/apps"), path)
}

func TestResolveAppsPathRequiresSplunkHomeForDefault(t *testing.T) {
	t.Setenv("SPLUNK_HOME", "")

	_, err := resolveAppsPath(defaultAppsPath)

	require.ErrorIs(t, err, errMissingSplunkHome)
}

type recordingNotify struct {
	id      observer.NotifyID
	mu      sync.Mutex
	added   []observer.Endpoint
	removed []observer.Endpoint
	changed []observer.Endpoint
}

func (n *recordingNotify) ID() observer.NotifyID {
	return n.id
}

func (n *recordingNotify) OnAdd(added []observer.Endpoint) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.added = append(n.added, added...)
}

func (n *recordingNotify) OnRemove(removed []observer.Endpoint) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.removed = append(n.removed, removed...)
}

func (n *recordingNotify) OnChange(changed []observer.Endpoint) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.changed = append(n.changed, changed...)
}

func (n *recordingNotify) hasAdded(id observer.EndpointID) bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	return hasEndpoint(n.added, id)
}

func (n *recordingNotify) hasRemoved(id observer.EndpointID) bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	return hasEndpoint(n.removed, id)
}

func hasEndpoint(endpoints []observer.Endpoint, id observer.EndpointID) bool {
	for _, endpoint := range endpoints {
		if endpoint.ID == id {
			return true
		}
	}
	return false
}
