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

package tomcat

var defaultMBeanYAML = `
tomcat-threadpool:
  objectName: "Catalina:type=ThreadPool,*"
  instanceFrom:
    - name
  values:
  - type: gauge
    instancePrefix: tomcat.ThreadPool.maxThreads
    attribute: maxThreads
  - type: gauge
    instancePrefix: tomcat.ThreadPool.currentThreadsBusy
    attribute: currentThreadsBusy

tomcat-utilityexecutor:
    objectName: "Catalina:type=UtilityExecutor"
    values:
    - type: gauge
      instancePrefix: tomcat.UtilityExecutor.activeCount
      attribute: activeCount
    - type: gauge
      instancePrefix: tomcat.UtilityExecutor.maximumPoolSize
      attribute: maximumPoolSize

tomcat-globalrequestprocessor:
    objectName: "Catalina:type=GlobalRequestProcessor,*"
    instanceFrom:
      - name
    values:
    - type: counter
      instancePrefix: tomcat.GlobalRequestProcessor.bytesSent
      attribute: bytesSent
    - type: counter
      instancePrefix: tomcat.GlobalRequestProcessor.bytesReceived
      attribute: bytesReceived
    - type: counter
      instancePrefix: tomcat.GlobalRequestProcessor.errorCount
      attribute: errorCount
    - type: counter
      instancePrefix: tomcat.GlobalRequestProcessor.requestCount
      attribute: requestCount
    - type: gauge
      instancePrefix: tomcat.GlobalRequestProcessor.maxTime
      attribute: maxTime
    - type: counter
      instancePrefix: tomcat.GlobalRequestProcessor.processingTime
      attribute: processingTime
`
