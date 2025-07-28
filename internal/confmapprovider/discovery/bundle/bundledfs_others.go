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

//go:build !windows

package bundle

import "embed"

// BundledFS is the in-executable filesystem that contains all bundled discovery config.d components.

//go:embed bundle.d/extensions/docker-observer.discovery.yaml
//go:embed bundle.d/extensions/host-observer.discovery.yaml
//go:embed bundle.d/extensions/k8s-observer.discovery.yaml
//go:embed bundle.d/receivers/apache.discovery.yaml
//go:embed bundle.d/receivers/envoy.discovery.yaml
//go:embed bundle.d/receivers/istio.discovery.yaml
//go:embed bundle.d/receivers/jmx-cassandra.discovery.yaml
//go:embed bundle.d/receivers/kafkametrics.discovery.yaml
//go:embed bundle.d/receivers/mongodb.discovery.yaml
//go:embed bundle.d/receivers/mysql.discovery.yaml
//go:embed bundle.d/receivers/nginx.discovery.yaml
//go:embed bundle.d/receivers/oracledb.discovery.yaml
//go:embed bundle.d/receivers/postgresql.discovery.yaml
//go:embed bundle.d/receivers/rabbitmq.discovery.yaml
//go:embed bundle.d/receivers/redis.discovery.yaml
//go:embed bundle.d/receivers/sqlserver.discovery.yaml
var BundledFS embed.FS
