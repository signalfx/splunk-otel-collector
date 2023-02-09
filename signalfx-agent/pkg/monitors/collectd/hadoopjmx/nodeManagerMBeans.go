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

//go:build linux
// +build linux

package hadoopjmx

var defaultNodeManagerMBeanYAML = `
hadoop-nodeManager-metrics:
  objectName: "Hadoop:name=NodeManagerMetrics,service=NodeManager"
  values:
  - instancePrefix: "hadoop-nodeManager-allocated-memory"
    type: "gauge"
    table: false
    attribute: "AllocatedGB"
  - instancePrefix: "hadoop-nodeManager-available-memory"
    type: "gauge"
    table: false
    attribute: "AvailableGB"
  - instancePrefix: "hadoop-nodeManager-containers-launched"
    type: "counter"
    table: false
    attribute: "ContainersLaunched"
  - instancePrefix: "hadoop-nodeManager-containers-failed"
    type: "counter"
    table: false
    attribute: "ContainersFailed"
  - instancePrefix: "hadoop-nodeManager-allocated-vcores"
    type: "gauge"
    table: false
    attribute: "AllocatedVCores"
  - instancePrefix: "hadoop-nodeManager-available-vcores"
    type: "gauge"
    table: false
    attribute: "AvailableVCores"
`
