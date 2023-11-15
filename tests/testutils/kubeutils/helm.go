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

package kubeutils

import (
	"context"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
)

type TestConfig struct {
	KindCluster   *KindCluster
	Configuration *action.Configuration
	Settings      *cli.EnvSettings
}

type SettingsFunc func(settings *cli.EnvSettings)

func Helm(kind *KindCluster, opts ...SettingsFunc) TestConfig {
	settings := cli.New()
	settings.KubeConfig = kind.Kubeconfig
	for _, fn := range opts {
		fn(settings)
	}
	cfg := new(action.Configuration)
	err := cfg.Init(
		settings.RESTClientGetter(), settings.Namespace(), "memory",
		func(format string, v ...interface{}) {
			kind.Testcase.Logf(format, v...)
		},
	)
	require.NoError(kind.Testcase, err)

	tc := TestConfig{
		KindCluster:   kind,
		Configuration: cfg,
		Settings:      settings,
	}
	return tc
}

type InstallFunc func(install *action.Install)

func (tc TestConfig) Install(chartPath, values string, opts ...InstallFunc) (*release.Release, error) {
	t := tc.KindCluster.Testcase
	vals := map[string]any{}
	require.NoError(t, yaml.Unmarshal([]byte(values), &vals))
	client := action.NewInstall(tc.Configuration)
	client.GenerateName = true

	name, path, err := client.NameAndChart([]string{chartPath})
	require.NoError(t, err)
	client.ReleaseName = name

	cp, err := client.ChartPathOptions.LocateChart(path, tc.Settings)
	if err != nil {
		return nil, err
	}

	chart, err := loader.Load(cp)
	require.NoError(t, err)

	c, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	for _, fn := range opts {
		fn(client)
	}

	return client.RunWithContext(c, chart, vals)
}
