//go:build linux
// +build linux

package genericjmx

// MBeanMap is a map from the service name to the mbean definitions that this
// service has
type MBeanMap map[string]MBean

// MergeWith combines the current MBeanMap with the one given as an
// argument and returns a new map with values from both maps.
func (m MBeanMap) MergeWith(m2 MBeanMap) MBeanMap {
	out := MBeanMap{}
	for k, v := range m {
		out[k] = v
	}
	for k, v := range m2 {
		out[k] = v
	}
	return out
}

// MBeanNames returns a list of the MBean names (the key values of the map)
func (m MBeanMap) MBeanNames() []string {
	names := make([]string, 0)
	for n := range m {
		names = append(names, n)
	}
	return names
}

// DefaultMBeans are basic JVM memory and threading metrics that are common to
// all JMX applications
var DefaultMBeans MBeanMap

const defaultMBeanYAML = `
classes:
  objectName: java.lang:type=ClassLoading
  values:
  - type: gauge
    instancePrefix: loaded_classes
    table: false
    attribute: LoadedClassCount

garbage_collector:
  objectName: "java.lang:type=GarbageCollector,*"
  instancePrefix: "gc-"
  instanceFrom:
  - "name"
  values:
  - type: "invocations"
    table: false
    attribute: "CollectionCount"
  - type: "total_time_in_ms"
    instancePrefix: "collection_time"
    table: false
    attribute: "CollectionTime"

memory-heap:
  objectName: java.lang:type=Memory
  instancePrefix: memory-heap
  values:
  - type: jmx_memory
    table: true
    attribute: HeapMemoryUsage

memory-nonheap:
  objectName: java.lang:type=Memory
  instancePrefix: memory-nonheap
  values:
  - type: jmx_memory
    table: true
    attribute: NonHeapMemoryUsage

memory_pool:
  objectName: java.lang:type=MemoryPool,*
  instancePrefix: memory_pool-
  instanceFrom:
  - name
  values:
  - type: jmx_memory
    table: true
    attribute: Usage

threading:
  objectName: java.lang:type=Threading
  values:
  - type: gauge
    table: false
    instancePrefix: jvm.threads.count
    attribute: ThreadCount
`

// MBeanValue specifies a particular value to pull from the MBean.
type MBeanValue struct {
	// Sets the data set used within collectd to handle the values
	// of the MBean attribute
	Type string `yaml:"type"`
	// Set this to true if the returned attribute is a composite type.
	// If set to true, the keys within the composite type is appended
	// to the type instance.
	Table bool `yaml:"table"`
	// Works like the option of the same name directly beneath the
	// MBean block, but sets the type instance instead
	InstancePrefix string `yaml:"instancePrefix"`
	//  Works like the option of the same name directly beneath the
	// MBean block, but sets the type instance instead
	InstanceFrom []string `yaml:"instanceFrom"`
	// Sets the name of the attribute from which to read the value.
	// You can access the keys of composite types by using a dot to
	// concatenate the key name to the attribute name.
	// For example: “attrib0.key42”. If `table` is set to true, path
	// must point to a composite type, otherwise it must point to
	// a numeric type.
	Attribute string `yaml:"attribute"`
	// The plural form of the `attribute` config above.  Used to derive
	// multiple metrics from a single MBean.
	Attributes []string `yaml:"attributes"`
}

// MBean represents the <MBean> config object in the collectd config for
// generic jmx.
type MBean struct {
	// Sets the pattern which is used to retrieve MBeans from the MBeanServer.
	// If more than one MBean is returned you should use the `instanceFrom` option
	// to make the identifiers unique
	ObjectName string `yaml:"objectName"`
	// Prefixes the generated plugin instance with prefix
	InstancePrefix string `yaml:"instancePrefix"`
	// The object names used by JMX to identify MBeans include so called
	// "properties" which are basically key-value-pairs. If the given object
	// name is not unique and multiple MBeans are returned, the values of those
	// properties usually differ. You can use this option to build the plugin
	// instance from the appropriate property values.
	// This option is optional and may be repeated to generate the plugin
	// instance from multiple property values
	InstanceFrom []string `yaml:"instanceFrom"`
	// The `value` blocks map one or more attributes of an MBean to a value
	// list in collectd. There must be at least one `value` block within each MBean block
	Values     []MBeanValue `yaml:"values"`
	Dimensions []string     `yaml:"dimensions"`
}
