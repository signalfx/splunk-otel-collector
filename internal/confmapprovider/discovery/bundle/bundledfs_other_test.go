// Copyright  Splunk, Inc.
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

//go:build !windows

package bundle

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBundleDir(t *testing.T) {
	receivers, err := fs.Glob(BundledFS, "bundle.d/receivers/*.discovery.yaml")
	require.NoError(t, err)
	require.Equal(t, []string{
		"bundle.d/receivers/apache.discovery.yaml",
		"bundle.d/receivers/envoy.discovery.yaml",
		"bundle.d/receivers/istio.discovery.yaml",
		"bundle.d/receivers/jmx-cassandra.discovery.yaml",
		"bundle.d/receivers/kafkametrics.discovery.yaml",
		"bundle.d/receivers/mongodb.discovery.yaml",
		"bundle.d/receivers/mysql.discovery.yaml",
		"bundle.d/receivers/nginx.discovery.yaml",
		"bundle.d/receivers/oracledb.discovery.yaml",
		"bundle.d/receivers/postgresql.discovery.yaml",
		"bundle.d/receivers/rabbitmq.discovery.yaml",
		"bundle.d/receivers/redis.discovery.yaml",
		"bundle.d/receivers/sqlserver.discovery.yaml",
	}, receivers)

	extensions, err := fs.Glob(BundledFS, "bundle.d/extensions/*.discovery.yaml")
	require.NoError(t, err)
	require.Equal(t, []string{
		"bundle.d/extensions/docker-observer.discovery.yaml",
		"bundle.d/extensions/host-observer.discovery.yaml",
		"bundle.d/extensions/k8s-observer.discovery.yaml",
	}, extensions)
}
