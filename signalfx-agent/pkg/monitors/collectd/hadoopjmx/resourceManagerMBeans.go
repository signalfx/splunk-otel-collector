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

var defaultResourceManagerMBeanYAML = `
hadoop-resourceManager-jvm:
  objectName: "Hadoop:service=ResourceManager,name=JvmMetrics"
  values:
  - instancePrefix: "hadoop-resourceManager-heap-used"
    type: "gauge"
    table: false
    attribute: "MemHeapUsedM"
  - instancePrefix: "hadoop-resourceManager-heap-max"
    type: "gauge"
    table: false
    attribute: "MemHeapMaxM"

hadoop-resourceManager-cluster-metrics:
  objectName: "Hadoop:service=ResourceManager,name=ClusterMetrics"
  values:
  - instancePrefix: "hadoop-resourceManager-active-nms"
    type: "gauge"
    table: false
    attribute: "NumActiveNMs"
    
hadoop-resourceManager-queue-metrics:
  objectName: "Hadoop:service=ResourceManager,name=QueueMetrics,q0=root"
  values:
  - instancePrefix: "hadoop-resourceManager-active-apps"
    type: "gauge"
    table: false
    attribute: "ActiveApplications"
  - instancePrefix: "hadoop-resourceManager-available-memory"
    type: "gauge"
    table: false
    attribute: "AvailableMB"
  - instancePrefix: "hadoop-resourceManager-allocated-memory"
    type: "gauge"
    table: false
    attribute: "AllocatedMB"
  - instancePrefix: "hadoop-resourceManager-active-users"
    type: "gauge"
    table: false
    attribute: "ActiveUsers"
  - instancePrefix: "hadoop-resourceManager-available-vcores"
    type: "gauge"
    table: false
    attribute: "AvailableVCores"
  - instancePrefix: "hadoop-resourceManager-allocated-vcores"
    type: "gauge"
    table: false
    attribute: "AllocatedVCores"
  - instancePrefix: "hadoop-resourceManager-allocated-containers"
    type: "gauge"
    table: false
    attribute: "AllocatedContainers"
`
