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

//go:build bundle.d

// These are the discovery config component generating statements.
// In order to update run go generate -tags bundle.d ./...
// TODO: add a pre-bundle.d step that generates these generate directives for easier maintenance/management across platforms
//go:generate go install github.com/signalfx/splunk-otel-collector/internal/confmapprovider/discovery/bundle/cmd/discoverybundler

//go:generate discoverybundler -r -t bundle.d/receivers/smartagent-postgresql.discovery.yaml.tmpl
//go:generate discoverybundler -r -c -d ../../../../cmd/otelcol/config/collector/config.d.linux/receivers -t bundle.d/receivers/smartagent-postgresql.discovery.yaml.tmpl
//go:generate discoverybundler -r -t bundle.d/receivers/smartagent-collectd-mysql.discovery.yaml.tmpl
//go:generate discoverybundler -r -c -d ../../../../cmd/otelcol/config/collector/config.d.linux/receivers -t bundle.d/receivers/smartagent-collectd-mysql.discovery.yaml.tmpl
//go:generate discoverybundler -r -t bundle.d/receivers/smartagent-collectd-nginx.discovery.yaml.tmpl
//go:generate discoverybundler -r -c -d ../../../../cmd/otelcol/config/collector/config.d.linux/receivers -t bundle.d/receivers/smartagent-collectd-nginx.discovery.yaml.tmpl

//go:generate discoverybundler -r -t bundle.d/extensions/docker-observer.discovery.yaml.tmpl
//go:generate discoverybundler -r -c -d ../../../../cmd/otelcol/config/collector/config.d.linux/extensions -t bundle.d/extensions/docker-observer.discovery.yaml.tmpl
//go:generate discoverybundler -r -t bundle.d/extensions/host-observer.discovery.yaml.tmpl
//go:generate discoverybundler -r -c -d ../../../../cmd/otelcol/config/collector/config.d.linux/extensions -t bundle.d/extensions/host-observer.discovery.yaml.tmpl
//go:generate discoverybundler -r -t bundle.d/extensions/k8s-observer.discovery.yaml.tmpl
//go:generate discoverybundler -r -c -d ../../../../cmd/otelcol/config/collector/config.d.linux/extensions -t bundle.d/extensions/k8s-observer.discovery.yaml.tmpl

package bundle
