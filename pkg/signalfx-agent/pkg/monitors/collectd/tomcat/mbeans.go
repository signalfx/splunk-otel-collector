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
