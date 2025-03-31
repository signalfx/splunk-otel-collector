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

//go:build bundle.d

// These are the discovery config component generating statements.
// In order to update run go generate -tags bundle.d ./...
//go:generate discoverybundler --render --template bundle.d/extensions/docker-observer.discovery.yaml.tmpl
//go:generate discoverybundler --render --commented --dir ../../../../cmd/otelcol/config/collector/config.d.linux/extensions -t bundle.d/extensions/docker-observer.discovery.yaml.tmpl
//go:generate discoverybundler --render --template bundle.d/extensions/host-observer.discovery.yaml.tmpl
//go:generate discoverybundler --render --commented --dir ../../../../cmd/otelcol/config/collector/config.d.linux/extensions -t bundle.d/extensions/host-observer.discovery.yaml.tmpl
//go:generate discoverybundler --render --template bundle.d/extensions/k8s-observer.discovery.yaml.tmpl
//go:generate discoverybundler --render --commented --dir ../../../../cmd/otelcol/config/collector/config.d.linux/extensions -t bundle.d/extensions/k8s-observer.discovery.yaml.tmpl

//go:generate discoverybundler --render --template bundle.d/receivers/apache.discovery.yaml.tmpl
//go:generate discoverybundler --render --commented --dir ../../../../cmd/otelcol/config/collector/config.d.linux/receivers -t bundle.d/receivers/apache.discovery.yaml.tmpl
//go:generate discoverybundler --render --template bundle.d/receivers/envoy.discovery.yaml.tmpl
//go:generate discoverybundler --render --commented --dir ../../../../cmd/otelcol/config/collector/config.d.linux/receivers -t bundle.d/receivers/envoy.discovery.yaml.tmpl
//go:generate discoverybundler --render --template bundle.d/receivers/istio.discovery.yaml.tmpl
//go:generate discoverybundler --render --commented --dir ../../../../cmd/otelcol/config/collector/config.d.linux/receivers -t bundle.d/receivers/istio.discovery.yaml.tmpl
//go:generate discoverybundler --render --template bundle.d/receivers/jmx-cassandra.discovery.yaml.tmpl
//go:generate discoverybundler --render --commented --dir ../../../../cmd/otelcol/config/collector/config.d.linux/receivers -t bundle.d/receivers/jmx-cassandra.discovery.yaml.tmpl
//go:generate discoverybundler --render --template bundle.d/receivers/kafkametrics.discovery.yaml.tmpl
//go:generate discoverybundler --render --commented --dir ../../../../cmd/otelcol/config/collector/config.d.linux/receivers -t bundle.d/receivers/kafkametrics.discovery.yaml.tmpl
//go:generate discoverybundler --render --template bundle.d/receivers/mongodb.discovery.yaml.tmpl
//go:generate discoverybundler --render --commented --dir ../../../../cmd/otelcol/config/collector/config.d.linux/receivers -t bundle.d/receivers/mongodb.discovery.yaml.tmpl
//go:generate discoverybundler --render --template bundle.d/receivers/mysql.discovery.yaml.tmpl
//go:generate discoverybundler --render --commented --dir ../../../../cmd/otelcol/config/collector/config.d.linux/receivers -t bundle.d/receivers/mysql.discovery.yaml.tmpl
//go:generate discoverybundler --render --template bundle.d/receivers/nginx.discovery.yaml.tmpl
//go:generate discoverybundler --render --commented --dir ../../../../cmd/otelcol/config/collector/config.d.linux/receivers -t bundle.d/receivers/nginx.discovery.yaml.tmpl
//go:generate discoverybundler --render --template bundle.d/receivers/oracledb.discovery.yaml.tmpl
//go:generate discoverybundler --render --commented --dir ../../../../cmd/otelcol/config/collector/config.d.linux/receivers -t bundle.d/receivers/oracledb.discovery.yaml.tmpl
//go:generate discoverybundler --render --template bundle.d/receivers/postgresql.discovery.yaml.tmpl
//go:generate discoverybundler --render --commented --dir ../../../../cmd/otelcol/config/collector/config.d.linux/receivers -t bundle.d/receivers/postgresql.discovery.yaml.tmpl
//go:generate discoverybundler --render --template bundle.d/receivers/rabbitmq.discovery.yaml.tmpl
//go:generate discoverybundler --render --commented --dir ../../../../cmd/otelcol/config/collector/config.d.linux/receivers -t bundle.d/receivers/rabbitmq.discovery.yaml.tmpl
//go:generate discoverybundler --render --template bundle.d/receivers/redis.discovery.yaml.tmpl
//go:generate discoverybundler --render --commented --dir ../../../../cmd/otelcol/config/collector/config.d.linux/receivers -t bundle.d/receivers/redis.discovery.yaml.tmpl
//go:generate discoverybundler --render --template bundle.d/receivers/sqlserver.discovery.yaml.tmpl
//go:generate discoverybundler --render --commented --dir ../../../../cmd/otelcol/config/collector/config.d.linux/receivers -t bundle.d/receivers/sqlserver.discovery.yaml.tmpl

package bundle
