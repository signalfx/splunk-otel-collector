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

package main

import "go.opentelemetry.io/collector/component"

type componentParams struct {
	TemplateFile    string
	ComponentID     component.ID
	SupportsWindows bool
}

var (
	receivers = []componentParams{
		{
			ComponentID:     component.MustNewID("apache"),
			TemplateFile:    "apache",
			SupportsWindows: true,
		},
		{
			ComponentID:     component.MustNewID("prometheus"),
			TemplateFile:    "envoy",
			SupportsWindows: true,
		},
		{
			ComponentID:     component.MustNewIDWithName("prometheus", "istio"),
			TemplateFile:    "istio",
			SupportsWindows: true,
		},
		{
			ComponentID:     component.MustNewIDWithName("jmx", "cassandra"),
			TemplateFile:    "jmx-cassandra",
			SupportsWindows: true,
		},
		{
			ComponentID:     component.MustNewID("kafkametrics"),
			TemplateFile:    "kafkametrics",
			SupportsWindows: true,
		},
		{
			ComponentID:     component.MustNewID("mongodb"),
			TemplateFile:    "mongodb",
			SupportsWindows: true,
		},
		{
			ComponentID:     component.MustNewID("mysql"),
			TemplateFile:    "mysql",
			SupportsWindows: true,
		},
		{
			ComponentID:     component.MustNewID("nginx"),
			TemplateFile:    "nginx",
			SupportsWindows: false,
		},
		{
			ComponentID:     component.MustNewID("oracledb"),
			TemplateFile:    "oracledb",
			SupportsWindows: true,
		},
		{
			ComponentID:     component.MustNewID("postgresql"),
			TemplateFile:    "postgresql",
			SupportsWindows: true,
		},
		{
			ComponentID:     component.MustNewID("rabbitmq"),
			TemplateFile:    "rabbitmq",
			SupportsWindows: true,
		},
		{
			ComponentID:     component.MustNewID("redis"),
			TemplateFile:    "redis",
			SupportsWindows: true,
		},
		{
			ComponentID:     component.MustNewID("sqlserver"),
			TemplateFile:    "sqlserver",
			SupportsWindows: true,
		},
	}

	extensions = []componentParams{
		{
			ComponentID:     component.MustNewID("docker_observer"),
			TemplateFile:    "docker-observer",
			SupportsWindows: true,
		},
		{
			ComponentID:     component.MustNewID("host_observer"),
			TemplateFile:    "host-observer",
			SupportsWindows: true,
		},
		{
			ComponentID:     component.MustNewID("k8s_observer"),
			TemplateFile:    "k8s-observer",
			SupportsWindows: true,
		},
	}

	Components = struct {
		Extensions []componentParams
		Receivers  []componentParams
	}{
		Extensions: extensions,
		Receivers:  receivers,
	}
)
